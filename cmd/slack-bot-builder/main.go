package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/innogames/slack-bot/v2/bot/builder"
)

// builds a custom bot binary which includes all plugins of the "plugins" config section, see docs/plugins.md
func main() {
	os.Exit(run())
}

func run() int {
	opts := builder.Options{}
	var showVersion bool
	flag.StringVar(&opts.ConfigPath, "config", "config.yaml", "Path to config.yaml with the \"plugins\" section. Can be a directory which will load all '*.yaml' inside")
	flag.StringVar(&opts.Output, "output", "slack-bot", "Path of the built bot binary")
	flag.StringVar(&opts.CLIOutput, "cli-output", "", "Path of the built CLI binary (bot emulator in the terminal), only built when set")
	flag.StringVar(&opts.Core, "core", "", "Version (like v2.5.0) or local directory of the slack-bot core. Default: version of this builder")
	flag.StringVar(&opts.WorkDir, "workdir", "", "Directory of the generated Go module, keep it to reuse the go.sum. Default: temporary directory")
	flag.BoolVar(&opts.DryRun, "dry-run", false, "Only resolve the plugins and print the generated files, without building")
	flag.BoolVar(&showVersion, "version", false, "Print the slack-bot version of this builder")
	flag.Parse()

	if showVersion {
		version := builder.GetBuilderVersion()
		if version == "" {
			version = "unknown (local build)"
		}
		fmt.Println(version)

		return 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := builder.Build(ctx, opts); err != nil {
		fmt.Fprintf(os.Stderr, "slack-bot-builder: %s\n", err)
		return 1
	}

	return 0
}
