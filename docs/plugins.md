# Plugins

The slack-bot core contains the framework (Slack connection, command matching, config, storage, queue, crons, help...)
and the common commands. Additional commands, especially the ones with heavy dependencies or special use cases, are
**plugins**: Go modules which are compiled into the bot binary on demand.

- A plugin is a normal Go module with its own `go.mod`, so it can use any client library.
- The `plugins` section of the `config.yaml` defines which plugins are used, where they come from
  (Go module, git repository or local directory) and their config, like API credentials.
- The `slack-bot-builder` reads this config and builds a bot binary including all listed plugins. Go modules resolve
  the dependencies of the core and all plugins into one consistent set, so there are no version conflicts at runtime.
- Plugins have full access to the framework, exactly like the built-in commands.

The official bot binary and Docker image only contain the core. Use the `slack-bot-builder` to get a binary with plugins.

## Configuration

```yaml
plugins:
  # the key is the name of the plugin
  aws:
    source: github.com/innogames/slack-bot/plugins/aws   # Go module
    version: v1.0.0                                      # optional, default: latest version
    config:                                              # plugin specific config
      cloud_front:
        - id: E1234ABCDEF
          name: my-distribution

  deploy:
    source: https://gitlab.example.com/team/slack-bot-plugins.git   # git repository
    version: v2.1.0                                      # tag, branch or commit, default: default branch
    subdir: deploy                                       # directory of the plugin within the repository
    config:
      api_token: ${DEPLOY_API_TOKEN}                     # replaced by the environment variable

  my_plugin:
    source: ./plugins/my_plugin                          # local directory, relative to the config file

  weather:
    source: github.com/innogames/slack-bot/plugins/weather
    enabled: false                                       # compiled in, but not loaded
```

| Option    | Used by          | Description                                                                                                    |
|-----------|------------------|----------------------------------------------------------------------------------------------------------------|
| `source`  | builder          | Go module path, git repository (`https://`, `ssh://`, `git@host:path`, `file://`) or local directory (`./`, `../`, `/`) |
| `version` | builder          | Version of the Go module or git ref (tag, branch, commit)                                                      |
| `subdir`  | builder          | Directory of the plugin package within the source, e.g. for a repository with multiple plugins                  |
| `package` | builder          | Go import path of the plugin package, only needed when it can't be detected from the source                     |
| `enabled` | bot              | `false` disables a compiled-in plugin without rebuilding the binary. Default: `true`                            |
| `config`  | bot (the plugin) | Plugin specific config. `${ENV_VAR}` placeholders are replaced by the environment variable                     |

At startup the bot only loads the compiled-in plugins which are listed in the `plugins` section, and logs which plugins
are compiled in, enabled or disabled. A listed plugin which is not compiled into the binary is logged as error.

Like the rest of the config, the `plugins` section can be split into multiple files, e.g. the sources in the
`config.yaml` of a repository and the credentials in a separate `secrets.yaml` (start the bot with `-config <directory>`).
Values which are defined in a yaml file can also be overwritten via environment, e.g. `BOT_PLUGINS_DEPLOY_CONFIG_API_TOKEN`.

## Building a bot with plugins

The builder only needs Go (and git for git sources). Run it within the directory of your config:

```bash
go run github.com/innogames/slack-bot/v2/cmd/slack-bot-builder@latest -config config.yaml -output ./slack-bot
./slack-bot -config config.yaml
```

The builder uses its own version for the slack-bot core, so pin the version for reproducible builds, e.g.
`slack-bot-builder@v2.5.0`.

| Flag          | Description                                                                                                         |
|---------------|---------------------------------------------------------------------------------------------------------------------|
| `-config`     | Config file or directory with the `plugins` section. Default: `config.yaml`                                          |
| `-output`     | Path of the built bot binary. Default: `slack-bot`                                                                  |
| `-cli-output` | Also build the [CLI tool](../readme.md#cli-tool) with the same plugins, to test them in the terminal                  |
| `-core`       | Version (like `v2.5.0`) or local directory of the slack-bot core. Default: version of the builder                   |
| `-workdir`    | Keep the generated Go module in this directory, e.g. to inspect or cache it. Default: temporary directory           |
| `-dry-run`    | Only resolve the plugins and print the generated `go.mod` and `main.go`                                             |

The builder generates a small Go module: a `main.go` which imports the core and all plugins, and a `go.mod` which
requires them. Local directories and git repositories are added via `replace` directives. Then Go resolves all
dependencies (`go get`, `go mod tidy`) and builds the binary (`go build`). `GOOS`, `GOARCH`, `CGO_ENABLED`,
`GOPRIVATE`, `GOPROXY` etc. are passed through to Go.

### Versions and dependencies

- Go picks the minimal version of each dependency which satisfies the core and all plugins. Newer versions of a
  dependency of a plugin never sneak into a build, as long as the versions of the plugins are pinned.
- The version of the core is fixed: if a plugin requires a newer slack-bot version, the build fails with a hint
  to use a newer `-core` version or an older plugin version.
- `go version -m ./slack-bot` shows the exact versions of the core, the plugins and all dependencies of a binary.

### Private repositories

- Go modules: set `GOPRIVATE=gitlab.example.com` and provide the git credentials (e.g. via `~/.netrc` or an ssh key),
  see [Go modules reference](https://go.dev/ref/mod#private-modules). For GitLab subgroups use the `.git` suffix:
  `source: gitlab.example.com/group/subgroup/plugin.git`.
- Git repositories are cloned via the `git` command, so the usual git credentials (ssh agent, credential helper) are used.

### Team config repository

A typical setup is a repository per team or company, containing the config, own plugins and the build setup, see
[examples/custom_build](../examples/custom_build):

```
my-slack-bot/
  config.yaml          # including the "plugins" section
  plugins/my_plugin/   # own plugins, with go.mod
  Makefile             # "make build" runs the slack-bot-builder
  Dockerfile           # builds an image with the custom bot binary
```

## Official plugins

These plugins are part of this repository in [plugins/](../plugins), each one is its own Go module.

| Plugin      | Source                                           | Commands                                                                 |
|-------------|--------------------------------------------------|--------------------------------------------------------------------------|
| `aws`       | `github.com/innogames/slack-bot/plugins/aws`       | `aws cf list`, `aws cf clean <distribution> at <path>`, `ecs ls <cluster>`, `ecs restart <cluster> <service>` |
| `ripeatlas` | `github.com/innogames/slack-bot/plugins/ripeatlas` | `credits`, `traceroute <destination>`                                    |
| `weather`   | `github.com/innogames/slack-bot/plugins/weather`   | `weather`, `weather in <location>`                                       |
| `example`   | `github.com/innogames/slack-bot/plugins/example`   | `echo <text>`: minimal reference plugin                                  |

```yaml
plugins:
  aws:
    source: github.com/innogames/slack-bot/plugins/aws
    config:
      # AWS credentials are taken from the usual AWS environment variables and config files
      cloud_front:
        - id: E1234ABCDEF
          name: my-distribution
  ripeatlas:
    source: github.com/innogames/slack-bot/plugins/ripeatlas
    config:
      api_key: ${RIPE_ATLAS_API_KEY}
  weather:
    source: github.com/innogames/slack-bot/plugins/weather
    config:
      apikey: ${OPENWEATHER_API_KEY}
      location: "Hamburg, DE"
      units: metric
```

Within this repository, `make build/slack-bot-full` builds a bot including all official plugins
(see [plugins/all.yaml](../plugins/all.yaml)).

### Migration from older versions

The AWS, RIPE Atlas and weather commands were part of the core before. Their top-level config sections
(`aws`, `ripeatlas`, `open_weather`) are not used anymore, the bot logs a warning when they are still defined:

1. Add the plugin to the `plugins` section and move the old options into `plugins.<name>.config`
   (`aws.enabled` is not needed anymore: listing a plugin enables it).
2. Build the bot binary with the `slack-bot-builder`.

## Writing a plugin

A plugin is a Go package which registers itself in an `init()` function:

```go
package myplugin

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
)

// Config is defined in "plugins.my_plugin.config"
type Config struct {
	Greeting string `mapstructure:"greeting"`
}

type helloCommand struct {
	bot.BaseCommand
	cfg Config
}

func (c *helloCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("hello", func(_ matcher.Result, message msg.Message) {
		c.SendMessage(message, c.cfg.Greeting)
	})
}

func (c *helloCommand) GetHelp() []bot.Help {
	return []bot.Help{{Command: "hello", Description: "says hello"}}
}

func init() {
	bot.RegisterPlugin(bot.Plugin{
		Name: "my_plugin", // key in the "plugins" config section
		// only called when the plugin is enabled in the config
		Setup: func(ctx *bot.PluginContext) (bot.Commands, error) {
			commands := bot.Commands{}

			cfg := Config{Greeting: "Hello!"}
			if err := ctx.LoadConfig(&cfg); err != nil {
				return commands, err
			}

			commands.AddCommand(&helloCommand{ctx.BaseCommand(), cfg})

			return commands, nil
		},
	})
}
```

```
module gitlab.example.com/team/my-plugin

go 1.26.0

require github.com/innogames/slack-bot/v2 v2.5.0
```

If `Setup` returns an error (or panics), the error is logged and the bot starts without the plugin.
See [plugins/example](../plugins/example) for a complete plugin including tests.

### Framework access

Plugins use the framework exactly like the built-in commands in [command/](../command):

- **Commands and matchers**: implement `bot.Command` with `GetMatcher()`, using `TextMatcher`, `RegexpMatcher`
  (with named groups), `PrefixMatcher`, `GroupMatcher`, `AdminMatcher`... from `bot/matcher`.
- **Config**: `ctx.LoadConfig(&cfg)` decodes the `config` of the plugin, `ctx.Config` is the whole bot config.
- **Slack**: `ctx.SlackClient` (also embedded in `bot.BaseCommand`) to send messages, blocks, reactions, files...
  `ctx.Slack` is the full client, providing the whole Slack API (`*slack.Client`) and the socket mode client.
- **Storage**: `bot/storage` (`storage.Write`, `storage.Read`, `storage.Atomic`...) with the configured file/Redis backend.
  Use an own collection name, like the plugin name.
- **Queue**: `queue.AddRunningCommand()` from `command/queue`, to execute a command after a running process is done.
- **Background tasks**: implement `bot.Runnable` (`RunAsync(ctx *util.ServerContext)`) on a command, it's started
  with the bot and gets stopped on shutdown.
- **Crons**: plugin commands can be executed by the [crons](../readme.md#cron) of the config, own schedules can
  be implemented via `bot.Runnable`.
- **Help**: implement `bot.HelpProvider` (`GetHelp()`), so the commands show up in `help`.
- **Template functions**: implement `util.TemplateFunctionProvider`, to provide functions for custom commands and crons.
- **Conditional commands**: implement `bot.Conditional` (`IsEnabled()`).
- **Stats and metrics**: executions of plugin commands show up in `bot stats` automatically. Own counters via
  `stats.IncreaseOne()` from `bot/stats` are exported as Prometheus metrics (`slack_bot_<key>`),
  `stats.RegisterCollector()` adds own Prometheus collectors.
- **Logging**: `ctx.Logger` adds the plugin name to all log entries.

### Testing

`bot.NewPluginContext()` creates the context with a given config, `mocks.NewSlackClient()` mocks the Slack client:

```go
func TestPlugin(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	cfg := config.Config{
		Plugins: map[string]config.PluginConfig{
			"my_plugin": {Config: map[string]any{"greeting": "Moin"}},
		},
	}

	commands, err := setup(bot.NewPluginContext("my_plugin", slackClient, cfg))
	require.NoError(t, err)

	message := msg.Message{}
	message.Text = "hello"
	mocks.AssertSlackMessage(slackClient, message, "Moin")

	assert.True(t, commands.Run(message))
}
```

To try the plugin in the terminal, build the CLI tool including the plugin:
`slack-bot-builder -config config.yaml -cli-output ./cli && ./cli -config config.yaml`

### Developing core and plugin together

Use a local checkout of the slack-bot as core: `slack-bot-builder -core ../slack-bot ...`.
Within this repository, `make build-custom CONFIG=my-config.yaml` builds the plugins of the given config against
the local core.

## Development of the official plugins

- Each plugin in [plugins/](../plugins) is its own Go module (`github.com/innogames/slack-bot/plugins/<name>`),
  so its dependencies don't end up in the core module. The `replace` directive in its `go.mod` uses the core of this
  repository, it's ignored when the plugin is used as dependency.
- `make test`, `make test-race` and `make lint` also run in all plugin modules.
- Release: after releasing the core, update the required slack-bot version in the `go.mod` of the plugins
  and tag them as `plugins/<name>/vX.Y.Z` (e.g. `plugins/aws/v1.0.0`), so they can be used as Go modules.

## Why compile-time plugins?

Loading Go code at runtime comes with heavy trade-offs:

- **Go plugins** (`-buildmode=plugin`): the bot and each plugin must be built with the exact same Go version and the
  exact same version of every shared dependency, they need cgo and don't work on Windows.
- **RPC plugins** (like [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin)): the bot API is based on Go
  callbacks (matchers, Slack client, storage, queue...), every part would need an RPC protocol in both directions.
- **Interpreters and WASM**: no or only limited support for Go modules and the dependencies of plugins.

Compiling the plugins into the binary (like [xcaddy](https://github.com/caddyserver/xcaddy) for Caddy) keeps the full
type safety and framework access, and Go modules resolve all dependencies at build time.
