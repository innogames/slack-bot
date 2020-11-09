package config

// Config contains the full config structure of this bot
type Config struct {
	Slack       Slack     `mapstructure:"slack"`
	Server      Server    `mapstructure:"server"`
	Jenkins     Jenkins   `mapstructure:"jenkins"`
	Jira        Jira      `mapstructure:"jira"`
	Mqtt        Mqtt      `mapstructure:"mqtt"`
	StoragePath string    `mapstructure:"storage_path"`
	Bitbucket   Bitbucket `mapstructure:"bitbucket"`
	Github      struct {
		AccessToken string
	}
	Gitlab struct {
		AccessToken string
		Host        string
	}
	Macros []Macro `mapstructure:"macros"`
	Crons  []Cron  `mapstructure:"crons"`
	Logger Logger  `mapstructure:"logger"`

	// @deprecated
	BranchLookup struct {
		Type       string // stash/bitbucket/git/null
		Repository string
	} `mapstructure:"branch_lookup"`

	AllowedUsers []string    `mapstructure:"allowed_users,flow"`
	AdminUsers   []string    `mapstructure:"admin_users,flow"`
	OpenWeather  OpenWeather `mapstructure:"open_weather"`
	PullRequest  PullRequest `mapstructure:"pullrequest"`
	Timezone     string      `mapstructure:"timezone"`
}

// OpenWeather is an optional feature to get current weather
type OpenWeather struct {
	Apikey   string
	Location string
	Url      string
	Units    string
}

// Slack contains the credentials and configuration of the Slack client
type Slack struct {
	Token            string   `mapstructure:"token"`
	Debug            bool     `mapstructure:"debug"`
	AllowedGroups    []string `mapstructure:"allowed_groups,flow"`
	AutoJoinChannels []string `mapstructure:"auto_join_channels,flow"`
	ErrorChannel     string   `mapstructure:"error_channel"`

	// use websocket RTM connection.
	// this is NOT possible anymore for new apps!
	UseRTM      bool `mapstructure:"use_rtm"`
	UseEventAPI bool `mapstructure:"use_event_api"`

	// only used for integration tests
	TestEndpointUrl   string `mapstructure:"-"`
	VerificationToken string `mapstructure:"-"`
}

type Server struct {
	Listen        string `mapstructure:"listen"`
	SigningSecret string `mapstructure:"signing_secret"`
}

func (c Server) IsEnabled() bool {
	return c.Listen != "" && c.SigningSecret != ""
}

type Logger struct {
	Level string `mapstructure:"level"`
	File  string `mapstructure:"file"`
}

// Macro represents a single macro which is defined by a trigger regexp and a list of executed commands
type Macro struct {
	Name        string
	Description string
	Trigger     string
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

// Bitbucket credentials/options. Either add Username+Password OR a ApiKey
type Bitbucket struct {
	Host       string `mapstructure:"host"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	ApiKey     string `mapstructure:"api_key"`
	Project    string `mapstructure:"project"`
	Repository string `mapstructure:"repository"`
}

func (c Bitbucket) IsEnabled() bool {
	return c.Host != ""
}
