package config

// Config contains the full config structure of this bot
type Config struct {
	Slack       Slack
	Jenkins     Jenkins
	Jira        Jira
	Mqtt        Mqtt
	StoragePath string `yaml:"storage_path"`
	Bitbucket   Bitbucket
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

	AllowedUsers []string `yaml:"allowed_users,flow"`
	AdminUsers   []string `yaml:"admin_users,flow"`
}

// Slack contains the credentials and configuration of the Slack client
type Slack struct {
	Token             string
	Team              string
	Debug             bool
	AllowedGroups     []string `yaml:"allowed_groups,flow"`
	AutoJoinChannels  []string `yaml:"auto_join_channels,flow"`
	TestEndpointUrl   string
	VerificationToken string
}

type Logger struct {
	Level string
	File  string
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
