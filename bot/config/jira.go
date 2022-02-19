package config

// Jira configuration: credentials and custom formatting options
type Jira struct {
	Host        string
	Username    string
	Password    string
	AccessToken string
	Project     string
	Fields      []JiraField
}

// JiraField are custom Jira issue fields which should be displayed in the search/output
// Icons can be provided to have special mapping, e.g. for bug type or different priorities
type JiraField struct {
	Name  string
	Icons map[string]string
}

// IsEnabled checks if a host is defined (username/password) is not needed for public projects
func (c *Jira) IsEnabled() bool {
	return c.Host != ""
}
