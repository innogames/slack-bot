// Load application config, usually defined in config.yaml and overridable y env vars

package config

import (
	"github.com/brainexe/viper"
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
	Aws      Aws       `mapstructure:"aws"`
	Commands []Command `mapstructure:"commands"`
	Crons    []Cron    `mapstructure:"crons"`
	Logger   Logger    `mapstructure:"logger"`

	BranchLookup VCS `mapstructure:"branch_lookup"`

	// Metrics, like Prometheus
	Metrics Metrics `mapstructure:"metrics"`

	OpenWeather OpenWeather `mapstructure:"open_weather"`
	PullRequest PullRequest `mapstructure:"pullrequest"`
	Timezone    string      `mapstructure:"timezone"`

	// list of slack-bot plugins to load
	Plugins []string `mapstructure:"plugins"`

	// store whole Viper to get dynamic config values
	viper *viper.Viper `mapstructure:"-"`
}

// LoadCustom does a dynamic config lookup with a given key and unmarshals it into the value
func (c *Config) LoadCustom(key string, value any) error {
	if c.viper == nil {
		return nil
	}
	return c.viper.UnmarshalKey(key, value)
}

func (c *Config) Set(key string, value any) {
	if c.viper == nil {
		c.viper = viper.New()
	}
	c.viper.Set(key, value)
}

// Github config, currently just an access token
type Github struct {
	AccessToken string `mapstructure:"access_token"`
}

// OpenWeather is an optional feature to get current weather
type OpenWeather struct {
	Apikey   string
	Location string
	URL      string
	Units    string
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
	for _, userID := range l {
		if userID == givenUserID {
			return true
		}
	}

	return false
}

// UserMap indexed by user id, value is the user name
type UserMap map[string]string

// Contains checks if the given user is in the UserMap
func (m UserMap) Contains(givenUserID string) bool {
	return m[givenUserID] != ""
}
