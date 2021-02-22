package app

import (
	"embed"
	_ "embed"
	"strings"

	"github.com/gobuffalo/genny"
	"github.com/gobuffalo/plush"
	"github.com/gobuffalo/plushgen"
	"github.com/tendermint/starport/starport/pkg/cosmosver"
)

//go:embed app/templates/launchpad/*
var launchpad embed.FS

//go:embed app/templates/stargate/*
var stargate embed.FS

//Uses embed to embed stargate and Launchpad
var templates = map[cosmosver.MajorVersion]embed.FS{
	cosmosver.Launchpad: launchpad,
	cosmosver.Stargate: stargate,
}

// New ...
func New(sdkVersion cosmosver.MajorVersion, opts *Options) (*genny.Generator, error) {
	g := genny.New()
	if err := g.File(launchpad.ReadDir()
	ctx := plush.NewContext()
	ctx.Set("ModulePath", opts.ModulePath)
	ctx.Set("AppName", opts.AppName)
	ctx.Set("OwnerName", opts.OwnerName)
	ctx.Set("BinaryNamePrefix", opts.BinaryNamePrefix)
	ctx.Set("AddressPrefix", opts.AddressPrefix)
	ctx.Set("title", strings.Title)

	ctx.Set("nodash", func(s string) string {
		return strings.ReplaceAll(s, "-", "")
	})

	g.Transformer(plushgen.Transformer(ctx))
	g.Transformer(genny.Replace("{{appName}}", opts.AppName))
	g.Transformer(genny.Replace("{{binaryNamePrefix}}", opts.BinaryNamePrefix))
	return g, nil
}
