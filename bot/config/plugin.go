package config

// pluginKey returns the viper key of a plugin's config section, e.g. "plugins§aws" for "plugins: aws:"
func pluginKey(name string) string {
	return "plugins" + keyDelimiter + name
}

// LoadPlugin unmarshals a plugin's config from the "plugins: <name>:" section of the config.
// It falls back to the legacy top-level "<name>:" key (e.g. "aws:", "ripeatlas:")
// when no "plugins: <name>:" section is defined.
func (c *Config) LoadPlugin(name string, value any) error {
	if c.viper == nil {
		return nil
	}
	if c.viper.IsSet(pluginKey(name)) {
		return c.viper.UnmarshalKey(pluginKey(name), value)
	}
	return c.viper.UnmarshalKey(name, value)
}

// IsPluginEnabled is the global plugin kill switch: only an explicit
// "plugins: <name>: enabled: false" disables a plugin. A missing key means
// enabled - plugins can still self-gate on their own config, e.g. on missing credentials.
func (c *Config) IsPluginEnabled(name string) bool {
	if c.viper == nil {
		return true
	}
	key := pluginKey(name) + keyDelimiter + "enabled"
	if !c.viper.IsSet(key) {
		return true
	}
	return c.viper.GetBool(key)
}
