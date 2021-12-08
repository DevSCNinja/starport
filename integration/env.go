package envtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/starport/starport/chainconfig"
	"github.com/tendermint/starport/starport/pkg/availableport"
	"github.com/tendermint/starport/starport/pkg/cmdrunner"
	"github.com/tendermint/starport/starport/pkg/cmdrunner/step"
	"github.com/tendermint/starport/starport/pkg/gocmd"
	"github.com/tendermint/starport/starport/pkg/httpstatuschecker"
	"github.com/tendermint/starport/starport/pkg/xexec"
	"github.com/tendermint/starport/starport/pkg/xurl"
)

const (
	ServeTimeout = time.Minute * 15
	StarportApp  = "starport"
)

var isCI, _ = strconv.ParseBool(os.Getenv("CI"))

// Env provides an isolated testing environment and what's needed to
// make it possible.
type Env struct {
	t   *testing.T
	ctx context.Context
}

// New creates a new testing environment.
func New(t *testing.T) Env {
	ctx, cancel := context.WithCancel(context.Background())
	e := Env{
		t:   t,
		ctx: ctx,
	}
	t.Cleanup(cancel)

	if !xexec.IsCommandAvailable(StarportApp) {
		t.Fatal("starport needs to be installed")
	}

	return e
}

// SetCleanup registers a function to be called when the test (or subtest) and all its
// subtests complete.
func (e Env) SetCleanup(f func()) {
	e.t.Cleanup(f)
}

// Ctx returns parent context for the test suite to use for cancelations.
func (e Env) Ctx() context.Context {
	return e.ctx
}

type execOptions struct {
	ctx                    context.Context
	shouldErr, shouldRetry bool
	stdout, stderr         io.Writer
}

type ExecOption func(*execOptions)

// ExecShouldError sets the expectations of a command's execution to end with a failure.
func ExecShouldError() ExecOption {
	return func(o *execOptions) {
		o.shouldErr = true
	}
}

// ExecCtx sets cancelation context for the execution.
func ExecCtx(ctx context.Context) ExecOption {
	return func(o *execOptions) {
		o.ctx = ctx
	}
}

// ExecStdout captures stdout of an execution.
func ExecStdout(w io.Writer) ExecOption {
	return func(o *execOptions) {
		o.stdout = w
	}
}

// ExecStderr captures stderr of an execution.
func ExecStderr(w io.Writer) ExecOption {
	return func(o *execOptions) {
		o.stderr = w
	}
}

// ExecRetry retries command until it is successful before context is canceled.
func ExecRetry() ExecOption {
	return func(o *execOptions) {
		o.shouldRetry = true
	}
}

// Exec executes a command step with options where msg describes the expectation from the test.
// unless calling with Must(), Exec() will not exit test runtime on failure.
func (e Env) Exec(msg string, steps step.Steps, options ...ExecOption) (ok bool) {
	opts := &execOptions{
		ctx:    e.ctx,
		stdout: io.Discard,
		stderr: io.Discard,
	}
	for _, o := range options {
		o(opts)
	}
	var (
		stdout = &bytes.Buffer{}
		stderr = &bytes.Buffer{}
	)
	copts := []cmdrunner.Option{
		cmdrunner.DefaultStdout(io.MultiWriter(stdout, opts.stdout)),
		cmdrunner.DefaultStderr(io.MultiWriter(stderr, opts.stderr)),
	}
	if isCI {
		copts = append(copts, cmdrunner.EndSignal(os.Kill))
	}
	err := cmdrunner.
		New(copts...).
		Run(opts.ctx, steps...)
	if err == context.Canceled {
		err = nil
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if opts.shouldRetry && opts.ctx.Err() == nil {
			time.Sleep(time.Second)
			return e.Exec(msg, steps, options...)
		}
	}
	if err != nil {
		msg = fmt.Sprintf("%s\n\nLogs:\n\n%s\n\nError Logs:\n\n%s\n",
			msg,
			stdout.String(),
			stderr.String())
	}
	if opts.shouldErr {
		return assert.Error(e.t, err, msg)
	}
	return assert.NoError(e.t, err, msg)
}

const (
	Stargate = "stargate"
)

// Scaffold scaffolds an app to a unique appPath and returns it.
func (e Env) Scaffold(appName string, flags ...string) (appPath string) {
	root := e.TmpDir()
	e.Exec("scaffold an app",
		step.NewSteps(step.New(
			step.Exec(
				StarportApp,
				append([]string{
					"scaffold",
					"chain",
					fmt.Sprintf("github.com/test/%s", appName),
				}, flags...)...,
			),
			step.Workdir(root),
		)),
	)

	// Cleanup the home directory of the app
	e.t.Cleanup(func() {
		os.RemoveAll(filepath.Join(e.Home(), fmt.Sprintf(".%s", appName)))
	})

	return filepath.Join(root, appName)
}

// Serve serves an application lives under path with options where msg describes the
// execution from the serving action.
// unless calling with Must(), Serve() will not exit test runtime on failure.
func (e Env) Serve(msg, path, home, configPath string, options ...ExecOption) (ok bool) {
	serveCommand := []string{
		"chain",
		"serve",
		"-v",
	}

	if home != "" {
		serveCommand = append(serveCommand, "--home", home)
	}
	if configPath != "" {
		serveCommand = append(serveCommand, "--config", configPath)
	}

	return e.Exec(msg,
		step.NewSteps(step.New(
			step.Exec(StarportApp, serveCommand...),
			step.Workdir(path),
		)),
		options...,
	)
}

// Simulate runs the simulation test for the app
func (e Env) Simulate(appPath string, numBlocks, blockSize int) {
	e.Exec("running the simulation tests",
		step.NewSteps(step.New(
			step.Exec(
				StarportApp,
				"chain",
				"simulate",
				"--numBlocks",
				strconv.Itoa(numBlocks),
				"--blockSize",
				strconv.Itoa(blockSize),
			),
			step.Workdir(appPath),
		)),
	)
}

// EnsureAppIsSteady ensures that app living at the path can compile and its tests
// are passing.
func (e Env) EnsureAppIsSteady(appPath string) {
	_, statErr := os.Stat(filepath.Join(appPath, "config.yml"))
	require.False(e.t, os.IsNotExist(statErr), "config.yml cannot be found")

	e.Exec("make sure app is steady",
		step.NewSteps(step.New(
			step.Exec(gocmd.Name(), "test", "./..."),
			step.Workdir(appPath),
		)),
	)
}

// IsAppServed checks that app is served properly and servers are started to listening
// before ctx canceled.
func (e Env) IsAppServed(ctx context.Context, host chainconfig.Host) error {
	checkAlive := func() error {
		ok, err := httpstatuschecker.Check(ctx, xurl.HTTP(host.API)+"/node_info")
		if err == nil && !ok {
			err = errors.New("app is not online")
		}
		return err
	}
	return backoff.Retry(checkAlive, backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx))
}

// TmpDir creates a new temporary directory.
func (e Env) TmpDir() (path string) {
	path, err := os.MkdirTemp("", "integration")
	require.NoError(e.t, err, "create a tmp dir")
	e.t.Cleanup(func() { os.RemoveAll(path) })
	return path
}

// RandomizeServerPorts randomizes server ports for the app at path, updates
// its config.yml and returns new values.
func (e Env) RandomizeServerPorts(path string, configFile string) chainconfig.Host {
	if configFile == "" {
		configFile = "config.yml"
	}

	// generate random server ports and servers list.
	ports, err := availableport.Find(6)
	require.NoError(e.t, err)

	genAddr := func(port int) string {
		return fmt.Sprintf("localhost:%d", port)
	}

	servers := chainconfig.Host{
		RPC:     genAddr(ports[0]),
		P2P:     genAddr(ports[1]),
		Prof:    genAddr(ports[2]),
		GRPC:    genAddr(ports[3]),
		GRPCWeb: genAddr(ports[4]),
		API:     genAddr(ports[5]),
	}

	// update config.yml with the generated servers list.
	configyml, err := os.OpenFile(filepath.Join(path, configFile), os.O_RDWR|os.O_CREATE, 0755)
	require.NoError(e.t, err)
	defer configyml.Close()

	var conf chainconfig.Config
	require.NoError(e.t, yaml.NewDecoder(configyml).Decode(&conf))

	conf.Host = servers
	require.NoError(e.t, configyml.Truncate(0))
	_, err = configyml.Seek(0, 0)
	require.NoError(e.t, err)
	require.NoError(e.t, yaml.NewEncoder(configyml).Encode(conf))

	return servers
}

// SetRandomHomeConfig sets in the blockchain config files generated temporary directories for home directories
func (e Env) SetRandomHomeConfig(path string, configFile string) {
	if configFile == "" {
		configFile = "config.yml"
	}

	// update config.yml with the generated temporary directories
	configyml, err := os.OpenFile(filepath.Join(path, configFile), os.O_RDWR|os.O_CREATE, 0755)
	require.NoError(e.t, err)
	defer configyml.Close()

	var conf chainconfig.Config
	require.NoError(e.t, yaml.NewDecoder(configyml).Decode(&conf))

	conf.Init.Home = e.TmpDir()
	require.NoError(e.t, configyml.Truncate(0))
	_, err = configyml.Seek(0, 0)
	require.NoError(e.t, err)
	require.NoError(e.t, yaml.NewEncoder(configyml).Encode(conf))
}

// Must fails the immediately if not ok.
// t.Fail() needs to be called for the failing tests before running Must().
func (e Env) Must(ok bool) {
	if !ok {
		e.t.FailNow()
	}
}

// Home returns user's home dir.
func (e Env) Home() string {
	home, err := os.UserHomeDir()
	require.NoError(e.t, err)
	return home
}

// AppdHome returns appd's home dir.
func (e Env) AppdHome(name string) string {
	return filepath.Join(e.Home(), fmt.Sprintf(".%s", name))
}
