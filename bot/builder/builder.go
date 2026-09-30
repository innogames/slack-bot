// Package builder compiles a custom slack-bot binary which includes all plugins listed in the "plugins" config section.
//
// It generates a Go module with a main package importing the slack-bot core and the plugins, which are resolved by
// Go modules into one consistent set of dependencies, and runs "go build". See docs/plugins.md
package builder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/config"
)

// CoreModule is the module path of the slack-bot core
const CoreModule = "github.com/innogames/slack-bot/v2"

// Options of a custom bot build
type Options struct {
	// ConfigPath is the config file or directory (like for the bot itself) which contains the "plugins" section
	ConfigPath string

	// Output is the path of the built bot binary
	Output string

	// CLIOutput is the path of the built CLI binary (bot emulator in the terminal), it's only built when set
	CLIOutput string

	// Core is the version (like "v2.5.0") or a local directory of the slack-bot core. Default: version of the builder
	Core string

	// WorkDir is the directory of the generated Go module. Default: temporary directory, which is removed after the build
	WorkDir string

	// Tags is a comma separated list of Go build tags, e.g. "pprof"
	Tags string

	// DryRun only resolves the plugins and prints the generated files, without building the binaries
	DryRun bool

	// Log receives the progress and the output of the executed go and git commands. Default: os.Stderr
	Log io.Writer
}

type builder struct {
	opts        Options
	log         io.Writer
	baseDir     string            // relative plugin sources are based on the directory of the config
	workDir     string            // generated Go module
	coreVersion string            // used version of the slack-bot core
	clones      map[string]string // cloned git repositories -> commit
}

// Build compiles a custom bot binary with all plugins of the config
func Build(ctx context.Context, opts Options) error {
	b := &builder{
		opts:   opts,
		log:    opts.Log,
		clones: map[string]string{},
	}
	if b.log == nil {
		b.log = os.Stderr
	}
	if b.opts.Output == "" {
		b.opts.Output = "slack-bot"
	}

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config %s: %w", opts.ConfigPath, err)
	}

	b.baseDir, err = getBaseDir(opts.ConfigPath)
	if err != nil {
		return err
	}

	core, err := resolveCore(opts.Core)
	if err != nil {
		return err
	}

	if err = b.initWorkDir(); err != nil {
		return err
	}
	if opts.WorkDir == "" {
		defer os.RemoveAll(b.workDir)
	}

	plugins := make([]*plugin, 0, len(cfg.Plugins))
	for _, name := range slices.Sorted(maps.Keys(cfg.Plugins)) {
		p, err := b.resolvePlugin(ctx, name, cfg.Plugins[name])
		if err != nil {
			return fmt.Errorf("plugin %s: %w", name, err)
		}
		b.logf("Plugin %s: %s", name, p.describe())
		plugins = append(plugins, p)
	}
	if len(plugins) == 0 {
		b.logf(`No plugins defined in the "plugins" config section: building the slack-bot core only`)
	}

	if err = b.initModule(ctx, core); err != nil {
		return err
	}
	if err = b.addLocalPlugins(ctx, plugins); err != nil {
		return err
	}
	if err = b.writeMainPackages(plugins); err != nil {
		return err
	}

	if opts.DryRun {
		return b.printGeneratedFiles()
	}

	if err = b.resolveDependencies(ctx, core, plugins); err != nil {
		return err
	}

	if err = b.build(ctx, "./cmd/bot", b.opts.Output); err != nil {
		return err
	}
	if opts.CLIOutput != "" {
		if err = b.build(ctx, "./cmd/cli", opts.CLIOutput); err != nil {
			return err
		}
	}

	return nil
}

// coreSource is either a version (Go module query, like "v2.5.0" or "latest") or a local directory of the slack-bot
type coreSource struct {
	version string
	dir     string
}

func resolveCore(core string) (coreSource, error) {
	if core == "" {
		version := GetBuilderVersion()
		if version == "" {
			return coreSource{}, errors.New("can't detect the slack-bot version of this builder: use -core with a version (like v2.5.0) or a local directory of the slack-bot")
		}

		return coreSource{version: version}, nil
	}

	if isLocalPath(core) {
		dir, err := filepath.Abs(core)
		if err != nil {
			return coreSource{}, err
		}

		moduleDir, modulePath, err := findModule(dir, dir)
		if err != nil || modulePath != CoreModule {
			return coreSource{}, fmt.Errorf("%s is not a slack-bot directory", core)
		}

		return coreSource{dir: moduleDir}, nil
	}

	return coreSource{version: core}, nil
}

// GetBuilderVersion returns the slack-bot version of this builder, when it was installed like
// "go run github.com/innogames/slack-bot/v2/cmd/slack-bot-builder@v2.5.0". It's empty for local builds.
func GetBuilderVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Path != CoreModule {
		return ""
	}

	version := info.Main.Version
	if version == "" || version == "(devel)" || strings.HasSuffix(version, "+dirty") {
		return ""
	}

	return version
}

func getBaseDir(configPath string) (string, error) {
	stat, err := os.Stat(configPath)
	if err != nil {
		return "", err
	}

	if !stat.IsDir() {
		configPath = filepath.Dir(configPath)
	}

	return filepath.Abs(configPath)
}

func (b *builder) initWorkDir() error {
	var err error
	if b.opts.WorkDir == "" {
		b.workDir, err = os.MkdirTemp("", "slack-bot-builder-")
		return err
	}

	b.workDir, err = filepath.Abs(b.opts.WorkDir)
	if err != nil {
		return err
	}

	// git sources are always cloned again
	if err = os.RemoveAll(filepath.Join(b.workDir, "sources")); err != nil {
		return err
	}

	return os.MkdirAll(b.workDir, 0o750)
}

// resolveDependencies adds the requirements of the plugin modules and resolves all dependencies
func (b *builder) resolveDependencies(ctx context.Context, core coreSource, plugins []*plugin) error {
	var pluginArgs []string
	for _, p := range plugins {
		if p.kind == moduleSource {
			pluginArgs = append(pluginArgs, p.importPath+"@"+p.getVersion())
		}
	}

	if len(pluginArgs) > 0 {
		args := []string{"get"}
		if core.dir == "" {
			// passing the core version as well prevents an implicit upgrade of the core by a plugin
			args = append(args, CoreModule+"@"+b.coreVersion)
		}

		if err := b.run(ctx, b.workDir, "go", append(args, pluginArgs...)...); err != nil {
			if core.dir == "" {
				return fmt.Errorf("%w\nHint: a plugin might require a newer slack-bot core than %s: use a newer version via -core or an older version of the plugin", err, b.coreVersion)
			}

			return err
		}
	}

	if err := b.run(ctx, b.workDir, "go", "mod", "tidy"); err != nil {
		return err
	}

	if core.dir != "" {
		return nil
	}

	// local plugins might still require a newer core version
	selectedVersion, err := b.getSelectedCoreVersion(ctx)
	if err != nil {
		return err
	}
	if selectedVersion != b.coreVersion {
		return fmt.Errorf(
			"the plugins require slack-bot %s, but %s is requested (required by %s): use -core %s or older versions of the plugins",
			selectedVersion,
			b.coreVersion,
			strings.Join(b.getCoreRequirers(ctx, selectedVersion), ", "),
			selectedVersion,
		)
	}

	return nil
}

// getLocalCoreVersion returns the version of a local slack-bot directory, based on the git tags
func (b *builder) getLocalCoreVersion(ctx context.Context, dir string) string {
	if dir == "" {
		return ""
	}

	version, err := b.output(ctx, dir, "git", "describe", "--tags")
	if err != nil || version == "" {
		return "devel"
	}

	return version
}

func (b *builder) getSelectedCoreVersion(ctx context.Context) (string, error) {
	return b.output(ctx, b.workDir, "go", "list", "-m", "-f", "{{.Version}}", CoreModule)
}

// getCoreRequirers returns all modules which require the given version of the core
func (b *builder) getCoreRequirers(ctx context.Context, version string) []string {
	graph, _ := b.output(ctx, b.workDir, "go", "mod", "graph")

	requirers := make([]string, 0)
	for line := range strings.Lines(graph) {
		requirer, required, ok := strings.Cut(strings.TrimSpace(line), " ")
		if ok && required == CoreModule+"@"+version && requirer != moduleName {
			requirers = append(requirers, requirer)
		}
	}

	return requirers
}

func (b *builder) build(ctx context.Context, pkg string, output string) error {
	output, err := filepath.Abs(output)
	if err != nil {
		return err
	}

	b.logf("Building %s", output)

	args := []string{"build", "-trimpath"}
	if b.opts.Tags != "" {
		args = append(args, "-tags="+b.opts.Tags)
	}
	args = append(args, "-ldflags=-s -w -X "+CoreModule+"/bot/version.Version="+b.coreVersion, "-o", output, pkg)

	return b.run(ctx, b.workDir, "go", args...)
}

func (b *builder) logf(format string, args ...any) {
	fmt.Fprintf(b.log, "[slack-bot-builder] "+format+"\n", args...)
}

// run executes the command, the output is written to the log
func (b *builder) run(ctx context.Context, dir string, name string, args ...string) error {
	b.logf("%s %s", name, strings.Join(args, " "))

	cmd := b.command(ctx, dir, name, args...)
	cmd.Stdout = b.log
	cmd.Stderr = b.log

	return wrapCommandError(cmd.Run(), name, args, nil)
}

// runSilent executes the command, the output is only returned in case of an error
func (b *builder) runSilent(ctx context.Context, dir string, name string, args ...string) error {
	cmd := b.command(ctx, dir, name, args...)
	out, err := cmd.CombinedOutput()

	return wrapCommandError(err, name, args, out)
}

// output executes the command and returns the trimmed stdout
func (b *builder) output(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := b.command(ctx, dir, name, args...)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	out, err := cmd.Output()

	return strings.TrimSpace(string(out)), wrapCommandError(err, name, args, stderr.Bytes())
}

func (b *builder) command(ctx context.Context, dir string, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204 -- only go and git are executed, based on the given config
	cmd.Dir = dir
	cmd.Env = append(
		os.Environ(),
		// don't use a go.work of a parent directory
		"GOWORK=off",
		// don't wait for git credentials
		"GIT_TERMINAL_PROMPT=0",
	)

	return cmd
}

func wrapCommandError(err error, name string, args []string, output []byte) error {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf("%s %s failed: %s", name, strings.Join(args, " "), err)
	if len(bytes.TrimSpace(output)) > 0 {
		message += "\n" + string(bytes.TrimSpace(output))
	}

	return errors.New(message)
}
