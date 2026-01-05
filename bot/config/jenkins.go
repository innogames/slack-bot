package config

import (
	"sort"
)

// Jenkins is the main Jenkins config, including credentials and the whitelisted jobs
type Jenkins struct {
	Host     string      `yaml:"host"`
	Username string      `yaml:"username"`
	Password string      `yaml:"password"`
	Jobs     JenkinsJobs `yaml:"jobs"`
}

// IsEnabled checks if a host was defined...by default it's not set
func (c Jenkins) IsEnabled() bool {
	return c.Host != ""
}

// JobConfig concrete job configuration -> only defined jobs are (re)startable
type JobConfig struct {
	Parameters []JobParameter `yaml:"parameters"`
	Trigger    string         `yaml:"trigger"`
	OnStart    []string       `yaml:"on_start"`
	OnSuccess  []string       `yaml:"on_success"`
	OnFailure  []string       `yaml:"on_failure"`
}

// JobParameter are defined build parameters per job
type JobParameter struct {
	Name    string `yaml:"name"`
	Default string `yaml:"default"`
	Type    string `yaml:"type"`
}

// JenkinsJobs is the list of all (whitelisted) Jenkins jobs
type JenkinsJobs map[string]JobConfig

// GetSortedNames get all defined job names, sorted by name
func (j JenkinsJobs) GetSortedNames() []string {
	jobNames := make([]string, 0, len(j))
	for jobName := range j {
		jobNames = append(jobNames, jobName)
	}
	sort.Strings(jobNames)

	return jobNames
}
