package config

// JenkinsJobs is the list of all (whitelisted) Jenkins jobs
type JenkinsJobs map[string]JobConfig

// Jenkins is the main Jenkins config, including credentials and the whitelisted jobs
type Jenkins struct {
	Host     string
	Username string
	Password string
	Jobs     JenkinsJobs
}

func (c Jenkins) IsEnabled() bool {
	return c.Host != ""
}

type JobParameter struct {
	Name    string
	Default string
	Type    string
}

type JobConfig struct {
	Parameters []JobParameter
	Trigger    string
	OnStart    []string
	OnSuccess  []string
	OnFailure  []string
}
