package scaffolder

import (
	"context"
	"fmt"
	"strings"

	"github.com/gobuffalo/genny"
	"github.com/tendermint/starport/starport/pkg/multiformatname"
	"github.com/tendermint/starport/starport/pkg/placeholder"
	"github.com/tendermint/starport/starport/pkg/xgenny"
	"github.com/tendermint/starport/starport/templates/field"
	"github.com/tendermint/starport/starport/templates/field/datatype"
	modulecreate "github.com/tendermint/starport/starport/templates/module/create"
	"github.com/tendermint/starport/starport/templates/typed"
	"github.com/tendermint/starport/starport/templates/typed/dry"
	"github.com/tendermint/starport/starport/templates/typed/list"
	maptype "github.com/tendermint/starport/starport/templates/typed/map"
	"github.com/tendermint/starport/starport/templates/typed/singleton"
)

// AddTypeOption configures options for AddType.
type AddTypeOption func(*addTypeOptions)

// AddTypeKind configures the type kind option for AddType.
type AddTypeKind func(*addTypeOptions)

type addTypeOptions struct {
	moduleName string
	fields     []string

	isList      bool
	isMap       bool
	isSingleton bool

	indexes []string

	withoutMessage bool
	signer         string
}

// newAddTypeOptions returns a addTypeOptions with default options
func newAddTypeOptions(moduleName string) addTypeOptions {
	return addTypeOptions{
		moduleName: moduleName,
		signer:     "creator",
	}
}

// ListType makes the type stored in a list convention in the storage.
func ListType() AddTypeKind {
	return func(o *addTypeOptions) {
		o.isList = true
	}
}

// MapType makes the type stored in a key-value convention in the storage with a custom
// index option.
func MapType(indexes ...string) AddTypeKind {
	return func(o *addTypeOptions) {
		o.isMap = true
		o.indexes = indexes
	}
}

// SingletonType makes the type stored in a fixed place as a single entry in the storage.
func SingletonType() AddTypeKind {
	return func(o *addTypeOptions) {
		o.isSingleton = true
	}
}

// DryType only creates a type with a basic definition.
func DryType() AddTypeKind {
	return func(o *addTypeOptions) {}
}

// TypeWithModule module to scaffold type into.
func TypeWithModule(name string) AddTypeOption {
	return func(o *addTypeOptions) {
		o.moduleName = name
	}
}

// TypeWithFields adds fields to the type to be scaffolded.
func TypeWithFields(fields ...string) AddTypeOption {
	return func(o *addTypeOptions) {
		o.fields = fields
	}
}

// TypeWithoutMessage disables generating sdk compatible messages and tx related APIs.
func TypeWithoutMessage() AddTypeOption {
	return func(o *addTypeOptions) {
		o.withoutMessage = true
	}
}

// TypeWithSigner provides a custom signer name for the message
func TypeWithSigner(signer string) AddTypeOption {
	return func(o *addTypeOptions) {
		o.signer = signer
	}
}

// AddType adds a new type to a scaffolded app.
// if non of the list, map or singleton given, a dry type without anything extra (like a storage layer, models, CLI etc.)
// will be scaffolded.
// if no module is given, the type will be scaffolded inside the app's default module.
func (s Scaffolder) AddType(
	ctx context.Context,
	typeName string,
	tracer *placeholder.Tracer,
	kind AddTypeKind,
	options ...AddTypeOption,
) (sm xgenny.SourceModification, err error) {
	// apply options.
	o := newAddTypeOptions(s.modpath.Package)
	for _, apply := range append(options, AddTypeOption(kind)) {
		apply(&o)
	}

	mfName, err := multiformatname.NewName(o.moduleName, multiformatname.NoNumber)
	if err != nil {
		return sm, err
	}
	moduleName := mfName.LowerCase

	name, err := multiformatname.NewName(typeName)
	if err != nil {
		return sm, err
	}

	if err := checkComponentValidity(s.path, moduleName, name, o.withoutMessage); err != nil {
		return sm, err
	}

	signer := ""
	if !o.withoutMessage {
		signer = o.signer
	}

	// Check and parse provided fields
	if err := checkCustomTypes(ctx, s.path, moduleName, o.fields); err != nil {
		return sm, err
	}
	tFields, err := field.ParseFields(o.fields, checkForbiddenTypeField, signer)
	if err != nil {
		return sm, err
	}

	mfSigner, err := multiformatname.NewName(o.signer)
	if err != nil {
		return sm, err
	}

	isIBC, err := isIBCModule(s.path, moduleName)
	if err != nil {
		return sm, err
	}

	var (
		g    *genny.Generator
		opts = &typed.Options{
			AppName:    s.modpath.Package,
			AppPath:    s.path,
			ModulePath: s.modpath.RawPath,
			ModuleName: moduleName,
			OwnerName:  owner(s.modpath.RawPath),
			TypeName:   name,
			Fields:     tFields,
			NoMessage:  o.withoutMessage,
			MsgSigner:  mfSigner,
			IsIBC:      isIBC,
		}
		gens []*genny.Generator
	)
	// Check and support MsgServer convention
	gens, err = supportMsgServer(
		gens,
		tracer,
		s.path,
		&modulecreate.MsgServerOptions{
			ModuleName: opts.ModuleName,
			ModulePath: opts.ModulePath,
			AppName:    opts.AppName,
			AppPath:    opts.AppPath,
			OwnerName:  opts.OwnerName,
		},
	)
	if err != nil {
		return sm, err
	}

	gens, err = supportGenesisTests(
		gens,
		opts.AppPath,
		opts.AppName,
		opts.ModulePath,
		opts.ModuleName,
	)
	if err != nil {
		return sm, err
	}

	gens, err = supportSimulation(
		gens,
		opts.AppPath,
		opts.ModulePath,
		opts.ModuleName,
	)
	if err != nil {
		return sm, err
	}

	gens, err = supportSimulation(
		gens,
		opts.AppPath,
		opts.ModulePath,
		opts.ModuleName,
	)
	if err != nil {
		return sm, err
	}

	// create the type generator depending on the model
	switch {
	case o.isList:
		g, err = list.NewStargate(tracer, opts)
	case o.isMap:
		g, err = mapGenerator(tracer, opts, o.indexes)
	case o.isSingleton:
		g, err = singleton.NewStargate(tracer, opts)
	default:
		g, err = dry.NewStargate(opts)
	}
	if err != nil {
		return sm, err
	}

	// run the generation
	gens = append(gens, g)
	sm, err = xgenny.RunWithValidation(tracer, gens...)
	if err != nil {
		return sm, err
	}

	return sm, finish(opts.AppPath, s.modpath.RawPath)
}

// checkForbiddenTypeIndex returns true if the name is forbidden as a field name
func checkForbiddenTypeIndex(name string) error {
	fieldSplit := strings.Split(name, datatype.Separator)
	if len(fieldSplit) > 1 {
		name = fieldSplit[0]
		fieldType := datatype.Name(fieldSplit[1])
		if f, ok := datatype.SupportedTypes[fieldType]; !ok || f.NonIndex {
			return fmt.Errorf("invalid index type %s", fieldType)
		}
	}
	return checkForbiddenTypeField(name)
}

// checkForbiddenTypeField returns true if the name is forbidden as a field name
func checkForbiddenTypeField(name string) error {
	mfName, err := multiformatname.NewName(name)
	if err != nil {
		return err
	}

	switch mfName.LowerCase {
	case
		"id",
		"params",
		"appendedvalue",
		datatype.TypeCustom:
		return fmt.Errorf("%s is used by type scaffolder", name)
	}

	return checkGoReservedWord(name)
}

// mapGenerator returns the template generator for a map
func mapGenerator(replacer placeholder.Replacer, opts *typed.Options, indexes []string) (*genny.Generator, error) {
	// Parse indexes with the associated type
	parsedIndexes, err := field.ParseFields(indexes, checkForbiddenTypeIndex)
	if err != nil {
		return nil, err
	}

	// Indexes and type fields must be disjoint
	exists := make(map[string]struct{})
	for _, name := range opts.Fields {
		exists[name.Name.LowerCamel] = struct{}{}
	}
	for _, index := range parsedIndexes {
		if _, ok := exists[index.Name.LowerCamel]; ok {
			return nil, fmt.Errorf("%s cannot simultaneously be an index and a field", index.Name.Original)
		}
	}

	opts.Indexes = parsedIndexes
	return maptype.NewStargate(replacer, opts)
}
