package config

// Config contains the full config structure of this bot
type Config struct {
	Slack       Slack     `mapstructure:"slack"`
	Server      Server    `mapstructure:"server"`
	Jenkins     Jenkins   `mapstructure:"jenkins"`
	Jira        Jira      `mapstructure:"jira"`
	StoragePath string    `mapstructure:"storage_path"`
	Bitbucket   Bitbucket `mapstructure:"bitbucket"`
	Github      struct {
		AccessToken string
	}
	Gitlab struct {
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

// OpenWeather is an optional feature to get current weather
type OpenWeather struct {
	Apikey   string
	Location string
	URL      string
	Units    string
}

// Slack contains the credentials and configuration of the Slack client
type Slack struct {
	Token            string   `mapstructure:"token"`
	AllowedGroups    []string `mapstructure:"allowed_groups,flow"`
	AutoJoinChannels []string `mapstructure:"auto_join_channels,flow"`
	ErrorChannel     string   `mapstructure:"error_channel"`

	Debug bool `mapstructure:"debug"`

	// use websocket RTM connection.
	// this is NOT possible anymore for new apps!
	UseRTM      bool `mapstructure:"use_rtm"`
	UseEventAPI bool `mapstructure:"use_event_api"`

	// only used for integration tests
	TestEndpointURL   string `mapstructure:"-"`
	VerificationToken string `mapstructure:"-"`
}

type Server struct {
	Listen        string `mapstructure:"listen"`
	SigningSecret string `mapstructure:"signing_secret"`
}

func (c Server) IsEnabled() bool {
	return c.Listen != "" && c.SigningSecret != ""
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

// Mqtt is a optional MQTT client to publish and subscribe values from the defined broker
type Mqtt struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func (c Mqtt) IsEnabled() bool {
	return c.Host != ""
}

// PullRequest special configuration to change the pull request behavior
type PullRequest struct {
	// able to set a custom "approved" reactions to see directly who or which component/department approved a pullrequest
	CustomApproveReaction map[string]string `mapstructure:"custom_approve_reaction"`
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

func (l UserList) Contains(userID string) bool {
	for _, adminID := range l {
		if adminID == userID {
			return true
		}
	}

	return false
}
