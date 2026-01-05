package config

// Jira configuration: credentials and custom formatting options
type Jira struct {
	Host        string      `yaml:"host"`
	Username    string      `yaml:"username"`
	Password    string      `yaml:"password"`
	AccessToken string      `yaml:"access_token"`
	Project     string      `yaml:"project"`
	Fields      []JiraField `yaml:"fields"`
}

// JiraField are custom Jira issue fields which should be displayed in the search/output
// Icons can be provided to have special mapping, e.g. for bug type or different priorities
type JiraField struct {
	Name  string            `yaml:"name"`
	Icons map[string]string `yaml:"icons"`
}

// IsEnabled checks if a host is defined (username/password) is not needed for public projects
func (c *Jira) IsEnabled() bool {
	return c.Host != ""
}
