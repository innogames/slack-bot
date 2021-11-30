// Load application config, usually defined in config.yaml and overrideable y env vars

package config

// Config contains the full config structure of this bot
type Config struct {
	Slack       Slack     `mapstructure:"slack"`
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

	// @deprecated
	BranchLookup struct {
		Type       string // stash/bitbucket/git/null
		Repository string
	} `mapstructure:"branch_lookup"`

	AllowedUsers UserList    `mapstructure:"allowed_users,flow"`
	AdminUsers   UserList    `mapstructure:"admin_users,flow"`
	OpenWeather  OpenWeather `mapstructure:"open_weather"`
	PullRequest  PullRequest `mapstructure:"pullrequest"`
	Timezone     string      `mapstructure:"timezone"`
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

// CanHandleInteractions checks if the slack config supports interaction/event via "Socket Mode" API
// in this case some commands are adding buttons to messages which are more advanced
func (s Slack) CanHandleInteractions() bool {
	return s.SocketToken != "" || s.IsFakeServer()
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
