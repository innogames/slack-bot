package config

import (
	"os"
	"regexp"
)

// PluginConfig is a single entry of the "plugins" config section, indexed by the plugin name.
//
// "source", "version", "subdir" and "package" are only used by the slack-bot-builder, which compiles the listed plugins
// into a custom bot binary. "enabled" and "config" are used by the bot at runtime.
type PluginConfig struct {
	// Source of the plugin: a Go module path (e.g. "github.com/innogames/slack-bot/plugins/aws"), a git repository
	// (e.g. "https://gitlab.example.com/team/slack-bot-plugins.git") or a local directory (e.g. "./plugins/my_plugin")
	Source string `mapstructure:"source"`

	// Version of the Go module or the git ref (tag, branch or commit) of a git repository.
	// Default: latest version of the module or the default branch of the git repository
	Version string `mapstructure:"version"`

	// Subdir is the directory of the plugin within the source, e.g. when a repository contains multiple plugins
	Subdir string `mapstructure:"subdir"`

	// Package is the Go import path of the plugin, only needed when it can't be detected by the source
	Package string `mapstructure:"package"`

	// Enabled can be set to "false" to disable a compiled-in plugin without rebuilding the bot binary
	Enabled *bool `mapstructure:"enabled"`

	// Config contains the plugin specific options, like API credentials. "${ENV_VAR}" placeholders get replaced
	Config map[string]any `mapstructure:"config"`
}

// IsEnabled checks if the plugin was not disabled explicitly
func (c PluginConfig) IsEnabled() bool {
	return c.Enabled == nil || *c.Enabled
}

// Decode the plugin specific "config" into the given target, "${ENV_VAR}" placeholders are replaced by the environment variable
func (c PluginConfig) Decode(target any) error {
	if c.Config == nil {
		return nil
	}

	return decode(expandEnv(c.Config), target)
}

var envPlaceholder = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// expandEnv replaces "${ENV_VAR}" placeholders in all (nested) string values. The given value is not modified.
func expandEnv(value any) any {
	switch v := value.(type) {
	case string:
		return envPlaceholder.ReplaceAllStringFunc(v, func(placeholder string) string {
			return os.Getenv(envPlaceholder.FindStringSubmatch(placeholder)[1])
		})
	case map[string]any:
		expanded := make(map[string]any, len(v))
		for key, val := range v {
			expanded[key] = expandEnv(val)
		}
		return expanded
	case []any:
		expanded := make([]any, len(v))
		for i, val := range v {
			expanded[i] = expandEnv(val)
		}
		return expanded
	default:
		return v
	}
}
