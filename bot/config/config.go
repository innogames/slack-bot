// Load application config, usually defined in config.yaml and overridable y env vars

package config

import (
	"slices"
	"strings"
)

// Config contains the full config structure of this bot
type Config struct {
	Slack Slack `mapstructure:"slack"`

	// authentication/authorization
	NoAuthentication bool     `mapstructure:"no_authentication"`
	AllowedUsers     UserList `mapstructure:"allowed_users,flow"`
	AdminUsers       UserList `mapstructure:"admin_users,flow"`

	Pool        Pool      `mapstructure:"pool"`
	Jenkins     Jenkins   `mapstructure:"jenkins"`
	Jira        Jira      `mapstructure:"jira"`
	StoragePath string    `mapstructure:"storage_path"`
	Bitbucket   Bitbucket `mapstructure:"bitbucket"`
	Github      Github    `mapstructure:"github"`
	Gitlab      struct {
		AccessToken string
		Host        string
	} `mapstructure:"gitlab"`
	Commands []Command `mapstructure:"commands"`
	Crons    []Cron    `mapstructure:"crons"`
	Logger   Logger    `mapstructure:"logger"`

	BranchLookup VCS `mapstructure:"branch_lookup"`

	// Metrics, like Prometheus
	Metrics Metrics `mapstructure:"metrics"`

	PullRequest PullRequest `mapstructure:"pullrequest"`
	Timezone    string      `mapstructure:"timezone"`

	// Plugins which are compiled into the bot binary (via slack-bot-builder), indexed by the plugin name
	Plugins map[string]PluginConfig `mapstructure:"plugins"`

	// store whole raw config to get dynamic config values
	raw map[string]any `mapstructure:"-"`
}

// LoadCustom does a dynamic config lookup with a given key (nested keys are separated by ".") and unmarshal it into the value
func (c *Config) LoadCustom(key string, value any) error {
	current, ok := c.lookup(key)
	if !ok || current == nil {
		return nil
	}

	return decode(current, value)
}

// IsSet checks if the given key (nested keys are separated by ".") is defined in the config
func (c *Config) IsSet(key string) bool {
	_, ok := c.lookup(key)

	return ok
}

func (c *Config) lookup(key string) (any, bool) {
	var current any = c.raw
	for part := range strings.SplitSeq(key, ".") {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		if current, ok = m[part]; !ok {
			return nil, false
		}
	}

	return current, true
}

// Set a dynamic config value (nested keys are separated by ".")...please only set it in tests!
func (c *Config) Set(key string, value any) {
	if c.raw == nil {
		c.raw = map[string]any{}
	}

	parts := strings.Split(key, ".")
	m := c.raw
	for _, part := range parts[:len(parts)-1] {
		subMap, ok := m[part].(map[string]any)
		if !ok {
			subMap = map[string]any{}
			m[part] = subMap
		}
		m = subMap
	}
	m[parts[len(parts)-1]] = value
}

// Github config, currently just an access token
type Github struct {
	AccessToken string `mapstructure:"access_token"`
}

// Slack contains the credentials and configuration of the Slack client
type Slack struct {
	Token         string   `mapstructure:"token"`
	SocketToken   string   `mapstructure:"socket_token"`
	AllowedGroups []string `mapstructure:"allowed_groups,flow"`
	ErrorChannel  string   `mapstructure:"error_channel"`

	Debug bool `mapstructure:"debug"`

	// only used for integration tests
	TestEndpointURL string `mapstructure:"-"`
}

// IsFakeServer is set for the "cli" tool which is spawning a fake test server which is mocking parts of the Slack API
func (s Slack) IsFakeServer() bool {
	return s.TestEndpointURL != ""
}

// Logger configuration to define log target or log levels
type Logger struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

// Command represents a single macro which is defined by a trigger regexp and a list of executed commands
type Command struct {
	Name        string
	Description string
	Trigger     string
	Category    string
	Commands    []string
	Examples    []string
}

// Bitbucket credentials/options. Either add Username+Password OR a APIKey
type Bitbucket struct {
	Host       string `mapstructure:"host"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	APIKey     string `mapstructure:"api_key"`
	Project    string `mapstructure:"project"`
	Repository string `mapstructure:"repository"`
}

// IsEnabled checks if a host is defined in the Bitbucket config
func (c *Bitbucket) IsEnabled() bool {
	return c.Host != ""
}

// UserList is a wrapper for []string with some helper to check is a user is in the list (e.g. used for AdminUser list)
type UserList []string

// Contains checks if the given user is in the UserList
func (l UserList) Contains(givenUserID string) bool {
	return slices.Contains(l, givenUserID)
}

// UserMap indexed by user id, value is the user name
type UserMap map[string]string

// Contains checks if the given user is in the UserMap
func (m UserMap) Contains(givenUserID string) bool {
	return m[givenUserID] != ""
}
