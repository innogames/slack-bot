package builder

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFile(t *testing.T, file string, content string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o750))
	require.NoError(t, os.WriteFile(file, []byte(content), 0o600))
}

func getCoreDir(t *testing.T) string {
	t.Helper()

	dir, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)

	return dir
}

func TestGetSourceKind(t *testing.T) {
	testCases := map[string]sourceKind{
		"./plugins/foo":                          localSource,
		"../plugins/foo":                         localSource,
		".":                                      localSource,
		"/opt/plugins/foo":                       localSource,
		"https://gitlab.example.com/team/p.git":  gitSource,
		"ssh://git@gitlab.example.com/team/p":    gitSource,
		"git@gitlab.example.com:team/p.git":      gitSource,
		"git+https://gitlab.example.com/team/p":  gitSource,
		"file:///opt/git/plugins":                gitSource,
		"github.com/innogames/slack-bot/plugins": moduleSource,
		"gitlab.example.com/team/sub/plugin.git": moduleSource,
	}

	for source, expected := range testCases {
		assert.Equal(t, expected, getSourceKind(source), source)
	}
}

func TestParseModulePath(t *testing.T) {
	assert.Equal(t, "example.com/foo", parseModulePath([]byte("module example.com/foo\n\ngo 1.26\n")))
	assert.Equal(t, "example.com/foo/v2", parseModulePath([]byte("// comment\nmodule \"example.com/foo/v2\" // comment\n")))
	assert.Empty(t, parseModulePath([]byte("go 1.26\n")))
}

func TestGetZeroVersion(t *testing.T) {
	assert.Equal(t, "v0.0.0-00010101000000-000000000000", getZeroVersion("example.com/foo"))
	assert.Equal(t, "v2.0.0-00010101000000-000000000000", getZeroVersion("github.com/innogames/slack-bot/v2"))
	assert.Equal(t, "v12.0.0-00010101000000-000000000000", getZeroVersion("example.com/foo/v12"))
	assert.Equal(t, "v0.0.0-00010101000000-000000000000", getZeroVersion("example.com/v2/foo"))
}

func TestFindModule(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "repo", "go.mod"), "module example.com/repo\n")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "repo", "plugins", "foo"), 0o750))

	moduleDir, modulePath, err := findModule(filepath.Join(dir, "repo", "plugins", "foo"), "")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "repo"), moduleDir)
	assert.Equal(t, "example.com/repo", modulePath)

	// the go.mod is outside of the given root
	_, _, err = findModule(filepath.Join(dir, "repo", "plugins", "foo"), filepath.Join(dir, "repo", "plugins"))
	require.ErrorContains(t, err, "no go.mod found in")
}

func TestResolveCore(t *testing.T) {
	// the version of the builder is not available in tests
	_, err := resolveCore("")
	require.ErrorContains(t, err, "can't detect the slack-bot version of this builder")

	core, err := resolveCore("v2.5.0")
	require.NoError(t, err)
	assert.Equal(t, coreSource{version: "v2.5.0"}, core)

	core, err = resolveCore("../..")
	require.NoError(t, err)
	assert.Equal(t, coreSource{dir: getCoreDir(t)}, core)

	_, err = resolveCore("./")
	require.EqualError(t, err, "./ is not a slack-bot directory")
}

func TestResolvePlugin(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "plugins", "go.mod"), "module example.com/plugins\n")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "plugins", "foo"), 0o750))

	b := &builder{baseDir: dir, workDir: t.TempDir(), clones: map[string]string{}}

	t.Run("local plugin", func(t *testing.T) {
		p, err := b.resolvePlugin(context.Background(), "foo", config.PluginConfig{Source: "./plugins", Subdir: "foo"})
		require.NoError(t, err)
		assert.Equal(t, localSource, p.kind)
		assert.Equal(t, "example.com/plugins", p.modulePath)
		assert.Equal(t, filepath.Join(dir, "plugins"), p.moduleDir)
		assert.Equal(t, "example.com/plugins/foo", p.importPath)
		assert.Equal(t, "package example.com/plugins/foo from local directory "+filepath.Join(dir, "plugins"), p.describe())
	})

	t.Run("module plugin", func(t *testing.T) {
		p, err := b.resolvePlugin(context.Background(), "aws", config.PluginConfig{Source: "github.com/innogames/slack-bot/plugins/aws"})
		require.NoError(t, err)
		assert.Equal(t, moduleSource, p.kind)
		assert.Equal(t, "github.com/innogames/slack-bot/plugins/aws", p.importPath)
		assert.Empty(t, p.moduleDir)
		assert.Equal(t, "package github.com/innogames/slack-bot/plugins/aws@latest", p.describe())

		p, err = b.resolvePlugin(context.Background(), "multi", config.PluginConfig{
			Source:  "gitlab.example.com/team/plugins",
			Subdir:  "deploy",
			Version: "v1.2.0",
		})
		require.NoError(t, err)
		assert.Equal(t, "gitlab.example.com/team/plugins/deploy", p.importPath)
		assert.Equal(t, "package gitlab.example.com/team/plugins/deploy@v1.2.0", p.describe())

		p, err = b.resolvePlugin(context.Background(), "custom", config.PluginConfig{
			Source:  "gitlab.example.com/team/plugins",
			Package: "gitlab.example.com/team/plugins/internal/custom",
		})
		require.NoError(t, err)
		assert.Equal(t, "gitlab.example.com/team/plugins/internal/custom", p.importPath)
	})

	t.Run("invalid plugins", func(t *testing.T) {
		testCases := map[string]config.PluginConfig{
			`no "source" defined`:                                                   {},
			`"source" and "version" must not start with "-"`:                        {Source: "--upload-pack=foo"},
			`"subdir" "../foo" must be a relative directory within the source`:      {Source: "./plugins", Subdir: "../foo"},
			`invalid source "plugins/foo": use a Go module path`:                    {Source: "plugins/foo"},
			"plugin directory " + filepath.Join(dir, "unknown") + " does not exist": {Source: "./unknown"},
		}

		for expectedError, cfg := range testCases {
			_, err := b.resolvePlugin(context.Background(), "invalid", cfg)
			require.ErrorContains(t, err, expectedError)
		}
	})
}

func TestWriteMain(t *testing.T) {
	file := filepath.Join(t.TempDir(), "main.go")

	plugins := []*plugin{
		{name: "aws", cfg: config.PluginConfig{Source: "github.com/innogames/slack-bot/plugins/aws"}, importPath: "github.com/innogames/slack-bot/plugins/aws"},
		{name: "deploy", cfg: config.PluginConfig{Source: "./plugins"}, importPath: "example.com/plugins"},
		{name: "rollback", cfg: config.PluginConfig{Source: "./plugins"}, importPath: "example.com/plugins"},
	}
	require.NoError(t, writeMain(file, "github.com/innogames/slack-bot/v2/bot/app", "app.Run()", plugins))

	expected := `// Code generated by slack-bot-builder. DO NOT EDIT.

package main

import (
	"github.com/innogames/slack-bot/v2/bot/app"

	// plugin "aws" (github.com/innogames/slack-bot/plugins/aws)
	_ "github.com/innogames/slack-bot/plugins/aws"

	// plugin "deploy" (./plugins), plugin "rollback" (./plugins)
	_ "example.com/plugins"
)

func main() {
	app.Run()
}
`
	content, err := os.ReadFile(file)
	require.NoError(t, err)
	assert.Equal(t, expected, string(content))

	// without plugins
	require.NoError(t, writeMain(file, "github.com/innogames/slack-bot/v2/bot/cli", "cli.Run()", nil))
	content, err = os.ReadFile(file)
	require.NoError(t, err)
	assert.Contains(t, string(content), "import (\n\t\"github.com/innogames/slack-bot/v2/bot/cli\"\n)\n")
}

func TestDryRun(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "plugins", "foo", "go.mod"), "module example.com/foo\n")
	writeFile(t, filepath.Join(dir, "config.yaml"), `
plugins:
  foo:
    source: ./plugins/foo
`)

	log := &bytes.Buffer{}
	err := Build(context.Background(), Options{
		ConfigPath: filepath.Join(dir, "config.yaml"),
		Core:       getCoreDir(t),
		CLIOutput:  "cli",
		DryRun:     true,
		Log:        log,
	})
	require.NoError(t, err)

	output := filepath.ToSlash(log.String())
	assert.Contains(t, output, "[slack-bot-builder] Plugin foo: package example.com/foo from local directory")
	assert.Contains(t, output, "// go.mod\nmodule slack-bot-custom")
	assert.Contains(t, output, "replace github.com/innogames/slack-bot/v2 => "+filepath.ToSlash(getCoreDir(t)))
	assert.Contains(t, output, "replace example.com/foo => "+filepath.ToSlash(filepath.Join(dir, "plugins", "foo")))
	assert.Contains(t, output, "// cmd/bot/main.go\n")
	assert.Contains(t, output, "// cmd/cli/main.go\n")
	assert.Contains(t, output, "\t_ \"example.com/foo\"\n")
}

// TestBuild builds a bot with a local and a git plugin and executes their commands in the built CLI
func TestBuild(t *testing.T) {
	if testing.Short() {
		t.Skip("skip building a bot binary in short mode")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()

	pluginSource := `package %PACKAGE%

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

type pingCommand struct {
	bot.BaseCommand
	reply string
}

func (c *pingCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("ping %PACKAGE%", func(_ matcher.Result, message msg.Message) {
		c.SendMessage(message, c.reply)
	})
}

func init() {
	bot.RegisterPlugin(bot.Plugin{
		Name: "%PACKAGE%",
		Setup: func(ctx *bot.PluginContext) (bot.Commands, error) {
			cfg := struct {
				Reply string
			}{Reply: "pong from %PACKAGE%"}
			err := ctx.LoadConfig(&cfg)

			commands := bot.Commands{}
			commands.AddCommand(&pingCommand{ctx.BaseCommand(), cfg.Reply})

			return commands, err
		},
	})
}
`
	goMod := "module %MODULE%\n\ngo 1.26.0\n\nrequire github.com/innogames/slack-bot/v2 v2.3.18\n"

	// local plugin in the config directory
	writeFile(t, filepath.Join(dir, "config", "plugins", "local", "go.mod"), strings.ReplaceAll(goMod, "%MODULE%", "example.com/local"))
	writeFile(t, filepath.Join(dir, "config", "plugins", "local", "local.go"), strings.ReplaceAll(pluginSource, "%PACKAGE%", "local"))

	// git repository with multiple plugins
	gitDir := filepath.Join(dir, "git")
	writeFile(t, filepath.Join(gitDir, "go.mod"), strings.ReplaceAll(goMod, "%MODULE%", "example.com/remote"))
	writeFile(t, filepath.Join(gitDir, "remote", "remote.go"), strings.ReplaceAll(pluginSource, "%PACKAGE%", "remote"))
	writeFile(t, filepath.Join(gitDir, "unused", "unused.go"), strings.ReplaceAll(pluginSource, "%PACKAGE%", "unused"))
	for _, args := range [][]string{
		{"init", "--quiet"},
		{"add", "."},
		{"-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "--quiet", "-m", "plugins"},
		{"tag", "v1.0.0"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = gitDir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}

	gitURL := filepath.ToSlash(gitDir)
	if !strings.HasPrefix(gitURL, "/") {
		gitURL = "/" + gitURL
	}

	writeFile(t, filepath.Join(dir, "config", "config.yaml"), `
plugins:
  local:
    source: ./plugins/local
    config:
      reply: "${BUILDER_TEST_REPLY}"
  remote:
    source: file://`+gitURL+`
    version: v1.0.0
    subdir: remote
`)

	botBinary := filepath.Join(dir, "slack-bot")
	cliBinary := filepath.Join(dir, "cli")
	if runtime.GOOS == "windows" {
		botBinary += ".exe"
		cliBinary += ".exe"
	}

	log := &bytes.Buffer{}
	err := Build(context.Background(), Options{
		ConfigPath: filepath.Join(dir, "config", "config.yaml"),
		Output:     botBinary,
		CLIOutput:  cliBinary,
		Core:       getCoreDir(t),
		WorkDir:    filepath.Join(dir, "work"),
		Log:        log,
	})
	require.NoError(t, err, log.String())

	assert.FileExists(t, botBinary)
	assert.FileExists(t, cliBinary)
	assert.Contains(t, log.String(), "[slack-bot-builder] Plugin remote: package example.com/remote/remote from file://")

	// the generated module is kept in the work dir
	goModContent, err := os.ReadFile(filepath.Join(dir, "work", "go.mod"))
	require.NoError(t, err)
	assert.Contains(t, string(goModContent), "example.com/local v0.0.0-00010101000000-000000000000")
	assert.Contains(t, string(goModContent), "example.com/remote v0.0.0-00010101000000-000000000000")

	// execute the plugin commands in the built CLI
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, cliBinary, "-config", filepath.Join(dir, "config", "config.yaml"))
	cmd.Env = append(os.Environ(), "BUILDER_TEST_REPLY=pong via config")
	output := &syncBuffer{}
	cmd.Stdout = output
	cmd.Stderr = output
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	_, err = stdin.Write([]byte("ping local\nping remote\nping unused\n"))
	require.NoError(t, err)

	found := assert.Eventually(t, func() bool {
		return strings.Contains(output.String(), "pong via config") &&
			strings.Contains(output.String(), "pong from remote") &&
			// the package with the "unused" plugin is not imported
			strings.Contains(output.String(), "ping unused") &&
			strings.Contains(output.String(), "not found...try")
	}, time.Second*30, time.Millisecond*100)
	if !found {
		t.Log(output.String())
	}
}

// TestBuildWithCoreVersion uses fake core and plugin modules in a local GOPROXY
func TestBuildWithCoreVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skip building a bot binary in short mode")
	}

	proxyDir := t.TempDir()
	coreFiles := map[string]string{
		"go.mod":         "module github.com/innogames/slack-bot/v2\n\ngo 1.21\n",
		"bot/app/app.go": "package app\n\n// Run is a fake of the bot\nfunc Run() {}\n",
		// only compiles when the build tag is passed to "go build"
		"bot/app/tag.go": "//go:build !fake_tag\n\npackage app\n\nvar _ = missingBuildTag\n",
	}
	publishModule(t, proxyDir, CoreModule, "v2.0.0", coreFiles)
	publishModule(t, proxyDir, CoreModule, "v2.1.0", coreFiles)
	publishModule(t, proxyDir, "example.com/plugin", "v1.0.0", map[string]string{
		"go.mod":    "module example.com/plugin\n\ngo 1.21\n\nrequire github.com/innogames/slack-bot/v2 v2.1.0\n",
		"plugin.go": "package plugin\n\nimport _ \"github.com/innogames/slack-bot/v2/bot/app\"\n",
	})

	proxyURL := filepath.ToSlash(proxyDir)
	if !strings.HasPrefix(proxyURL, "/") {
		proxyURL = "/" + proxyURL
	}
	t.Setenv("GOPROXY", "file://"+proxyURL)
	t.Setenv("GOSUMDB", "off")
	// the fake modules must not end up in the real module cache
	t.Setenv("GOMODCACHE", t.TempDir())
	t.Setenv("GOFLAGS", "-modcacherw")

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "local", "go.mod"), "module example.com/local\n\ngo 1.21\n\nrequire github.com/innogames/slack-bot/v2 v2.1.0\n")
	writeFile(t, filepath.Join(dir, "local", "local.go"), "package local\n")
	writeFile(t, filepath.Join(dir, "module.yaml"), "plugins:\n  plugin:\n    source: example.com/plugin\n    version: v1.0.0\n")
	writeFile(t, filepath.Join(dir, "local.yaml"), "plugins:\n  local:\n    source: ./local\n")

	build := func(configFile string, core string) (string, error) {
		log := &bytes.Buffer{}
		err := Build(context.Background(), Options{
			ConfigPath: filepath.Join(dir, configFile),
			Core:       core,
			Output:     filepath.Join(dir, "slack-bot"),
			Tags:       "fake_tag",
			Log:        log,
		})

		return log.String(), err
	}

	t.Run("latest core version", func(t *testing.T) {
		log, err := build("module.yaml", "latest")
		require.NoError(t, err, log)
		assert.Contains(t, log, "[slack-bot-builder] go get github.com/innogames/slack-bot/v2@v2.1.0 example.com/plugin@v1.0.0")
		assert.Contains(t, log, "[slack-bot-builder] go build -trimpath -tags=fake_tag -ldflags=-s -w -X github.com/innogames/slack-bot/v2/bot/version.Version=v2.1.0")
	})

	t.Run("module plugin requires a newer core", func(t *testing.T) {
		log, err := build("module.yaml", "v2.0.0")
		require.ErrorContains(t, err, "Hint: a plugin might require a newer slack-bot core than v2.0.0", log)
	})

	t.Run("local plugin requires a newer core", func(t *testing.T) {
		log, err := build("local.yaml", "v2.0.0")
		require.EqualError(
			t,
			err,
			"the plugins require slack-bot v2.1.0, but v2.0.0 is requested (required by example.com/local@v0.0.0-00010101000000-000000000000): "+
				"use -core v2.1.0 or older versions of the plugins",
			log,
		)
	})
}

// publishModule adds a version of a module to a GOPROXY directory
func publishModule(t *testing.T, proxyDir string, modulePath string, version string, files map[string]string) {
	t.Helper()

	dir := filepath.Join(proxyDir, filepath.FromSlash(modulePath), "@v")
	require.NoError(t, os.MkdirAll(dir, 0o750))

	list, err := os.OpenFile(filepath.Join(dir, "list"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = list.WriteString(version + "\n")
	require.NoError(t, err)
	require.NoError(t, list.Close())

	writeFile(t, filepath.Join(dir, version+".info"), `{"Version":"`+version+`","Time":"2026-01-01T00:00:00Z"}`)
	writeFile(t, filepath.Join(dir, version+".mod"), files["go.mod"])

	buf := &bytes.Buffer{}
	zipWriter := zip.NewWriter(buf)
	for name, content := range files {
		file, err := zipWriter.Create(modulePath + "@" + version + "/" + name)
		require.NoError(t, err)
		_, err = file.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, zipWriter.Close())
	require.NoError(t, os.WriteFile(filepath.Join(dir, version+".zip"), buf.Bytes(), 0o600))
}

type syncBuffer struct {
	buf  bytes.Buffer
	lock sync.Mutex
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.buf.String()
}
