package config

// Config contains the full config structure of this bot
type Config struct {
	Slack       Slack     `yaml:"slack"`
	Server      Server    `yaml:"server"`
	Jenkins     Jenkins   `yaml:"jenkins"`
	Jira        Jira      `yaml:"jira"`
	Mqtt        Mqtt      `yaml:"mqtt"`
	StoragePath string    `yaml:"storage_path"`
	Bitbucket   Bitbucket `yaml:"bitbucket"`
	Github      struct {
		AccessToken string
	}
	Gitlab struct {
		AccessToken string
		Host        string
	}
	Macros    []Macro
	Crons     []Cron
	Logger    Logger
	Calendars []Calendar

	// @deprecated
	BranchLookup struct {
		Type       string // stash/bitbucket/git/null
		Repository string
	} `yaml:"branch_lookup"`

	AllowedUsers []string    `yaml:"allowed_users,flow"`
	AdminUsers   []string    `yaml:"admin_users,flow"`
	OpenWeather  OpenWeather `yaml:"open_weather"`
	Timezone     string      `yaml:"timezone"`
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
	Token            string   `yaml:"token"`
	Debug            bool     `yaml:"debug"`
	AllowedGroups    []string `yaml:"allowed_groups,flow"`
	AutoJoinChannels []string `yaml:"auto_join_channels,flow"`
	ErrorChannel     string   `yaml:"error_channel"`

	// only used for integration tests
	TestEndpointUrl   string `yaml:"-"`
	VerificationToken string `yaml:"-"`
}

type Server struct {
	Listen string `yaml:"listen"`
}

type Logger struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
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
	Host     string
	Username string
	Password string
}

func (c Mqtt) IsEnabled() bool {
	return c.Host != ""
}

// Bitbucket credentials/options
type Bitbucket struct {
	Host       string
	Username   string
	Password   string
	Project    string
	Repository string
}

func (c Bitbucket) IsEnabled() bool {
	return c.Host != ""
}
