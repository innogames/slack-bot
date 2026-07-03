# Plugins

The slack-bot can be extended with **compile-time plugins**: self-contained Go modules which register
own commands into the bot framework. The heavy lifting (Slack connection, command matching, config,
storage, help system) stays in the core `github.com/innogames/slack-bot/v2` module — plugins and their
dependencies are only compiled into binaries which actually want them.

The `aws` and `ripeatlas` commands are shipped as plugins in [`plugins/`](../plugins) and are part of the
default `cmd/bot` binary and Docker image. The AWS SDK is a dependency of the
`plugins/aws` module only, not of the core slack-bot module anymore.

## Why compile-time plugins?

Runtime plugin loading in Go comes with heavy trade-offs: native `.so` plugins require the exact same
Go version and dependency versions in host and plugin, [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin)
requires an RPC-serializable API surface (the bot's matcher/handler API is callback based),
and interpreters like [yaegi](https://github.com/traefik/yaegi) don't support Go modules for plugin
dependencies. Compile-time composition (like [Caddy](https://github.com/caddyserver/xcaddy) does it)
keeps full type safety and full framework access — selecting plugins is a `go build` away.

## Using plugins

The default binary (`cmd/bot`, also used in the Docker image) ships with all in-repo plugins.
Plugins configure themselves from the `plugins:` section of `config.yaml`:

```yaml
plugins:
  aws:
    enabled: true
    cloud_front:
      - id: E1234ABCDEF
        name: my-distribution
  ripeatlas:
    api_key: "your-ripe-atlas-api-key"
```

- A compiled-in plugin can be disabled entirely with `plugins: <name>: enabled: false` —
  its `Init` is never executed.
- Without a `plugins:` section a plugin is loaded, but most plugins self-gate on their own
  config (e.g. the aws plugin requires `enabled: true`, ripeatlas requires an `api_key`).
- For backwards compatibility, the old top-level config keys (`aws:`, `ripeatlas:`) still work
  as long as no `plugins: <name>:` section is defined.

## Writing a plugin

A plugin is a Go package with an `init()` function which calls `bot.RegisterPlugin()`:

```go
package myplugin

import (
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
)

type Config struct {
	Token string `mapstructure:"token"`
}

type myCommand struct {
	bot.BaseCommand
	cfg Config
}

func (c *myCommand) GetMatcher() matcher.Matcher {
	return matcher.NewTextMatcher("my command", c.run)
}

func (c *myCommand) run(match matcher.Result, message msg.Message) {
	c.SendMessage(message, "hello from my plugin!")
}

func init() {
	bot.RegisterPlugin(bot.Plugin{
		Name: "myplugin",
		Init: func(slackClient client.SlackClient, cfg config.Config) bot.Commands {
			commands := bot.Commands{}

			// read the "plugins: myplugin:" config section
			var pluginCfg Config
			if err := cfg.LoadPlugin("myplugin", &pluginCfg); err != nil {
				return commands
			}
			if pluginCfg.Token == "" {
				return commands // self-gate: not configured -> no commands
			}

			commands.AddCommand(&myCommand{bot.BaseCommand{SlackClient: slackClient}, pluginCfg})
			return commands
		},
	})
}
```

See [`plugins/example`](../plugins/example) for a complete, tested reference plugin.

### What plugins can use

Everything a built-in command can:

- **Commands & matchers**: implement `bot.Command` with `GetMatcher()` — `TextMatcher`,
  `RegexpMatcher` (with named groups), `PrefixMatcher`, `GroupMatcher`, `AdminMatcher`, ...
- **Config**: `cfg.LoadPlugin("name", &pluginCfg)` reads the `plugins: name:` yaml section
  (with fallback to a legacy top-level `name:` key).
- **Slack API**: the passed `client.SlackClient` — `SendMessage`, `SendBlockMessage`,
  `AddReaction`, `GetThreadMessages`, file uploads etc.
- **Storage**: the package-global `bot/storage` key/value API (`storage.Write`, `storage.Read`,
  `storage.Atomic`, ...) with file/Redis/in-memory backends.
- **Background tasks**: implement `bot.Runnable` (`RunAsync(ctx *util.ServerContext)`) on a command —
  plugins are initialized before the async commands are started, so this works like for built-ins.
- **Help**: implement `bot.HelpProvider` (`GetHelp()`) so the commands show up in `help`.
- **Template functions**: implement `util.TemplateFunctionProvider` to extend custom commands/crons.

## In-repo plugins (this repository)

In-repo plugins live in `plugins/<name>/` as **own Go modules** (own `go.mod`), so their
dependencies don't leak into the core module. They are wired together by the
[`go.work`](../go.work) workspace and compiled into the default binaries via blank imports
in `cmd/bot/main.go` and `cmd/cli/main.go`.

Adding a new in-repo plugin:

1. Create `plugins/<name>/` with the plugin code and a `go.mod`
   (`module github.com/innogames/slack-bot/v2/plugins/<name>`, require `github.com/innogames/slack-bot/v2`
   plus your external dependencies).
2. Add `./plugins/<name>` to `go.work` and to `MODULE_DIRS` in the `Makefile`
   (plus the lint matrix in `.github/workflows/test.yaml`).
3. Add `_ "github.com/innogames/slack-bot/v2/plugins/<name>"` to the imports of
   `cmd/bot/main.go` and `cmd/cli/main.go` to ship it in the default binaries.
4. Run `make dep` to sync the vendor directory (`go work vendor`) and `make test`.

## External plugins (own repository)

Plugins can live in any repository — they are normal Go modules importing the slack-bot as library.
Build your own bot binary with your plugin selection:

```go
package main

import (
	"github.com/innogames/slack-bot/v2/bot/app"

	_ "github.com/mycompany/my-slack-bot-plugin"
	_ "github.com/innogames/slack-bot/v2/plugins/ripeatlas" // in-repo plugins work too
)

func main() {
	app.Run()
}
```

```
go mod init mycompany/my-bot && go mod tidy && go build -o my-bot .
```

Note: building against the in-repo plugin modules or the new plugin config API requires a
slack-bot release newer than v2.3.17.

## Limitations

- Plugins are selected at **compile time**. Changing the plugin set means rebuilding the binary
  (or Docker image); the `plugins:` config section only configures/disables compiled-in plugins.
- `BOT_*` environment variable overrides are not reliably applied to plugin config sections
  (pre-existing viper limitation of dynamic `UnmarshalKey` lookups).
- Plugin `Init` functions run once at startup (`bot.Init()`), before the Slack event loop starts.
