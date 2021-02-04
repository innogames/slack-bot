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
	}

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

func (s Slack) CanHandleInteractions() bool {
	return s.SocketToken != ""
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

func (c *Bitbucket) IsEnabled() bool {
	return c.Host != ""
}

type UserList []string

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

func (m UserMap) Contains(givenUserID string) bool {
	return m[givenUserID] != ""
}
