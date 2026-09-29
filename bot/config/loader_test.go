package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfigFile(t *testing.T, dir string, name string, content string) string {
	t.Helper()

	file := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(file, []byte(content), 0o600))

	return file
}

func TestLoadPreservesKeyCase(t *testing.T) {
	file := writeConfigFile(t, t.TempDir(), "config.yaml", `
jenkins:
  host: https://jenkins.example.com
  jobs:
    BuildApp:
      trigger: build app
    deploy-Service.Prod:
      parameters:
        - name: BRANCH
          default: master
jira:
  fields:
    - name: type
      icons:
        Story: ":book:"
pullrequest:
  custom_approve_reaction:
    JohnDoe: star
  jira_priority_reactions:
    Highest: fire
my_plugin:
  mapping:
    FooBar: baz
`)

	cfg, err := Load(file)
	require.NoError(t, err)

	assert.Equal(t, []string{"BuildApp", "deploy-Service.Prod"}, cfg.Jenkins.Jobs.GetSortedNames())
	assert.Equal(t, "build app", cfg.Jenkins.Jobs["BuildApp"].Trigger)
	assert.Equal(t, []JobParameter{{Name: "BRANCH", Default: "master"}}, cfg.Jenkins.Jobs["deploy-Service.Prod"].Parameters)

	require.Len(t, cfg.Jira.Fields, 1)
	// "Bug" is a default icon, see default.go
	assert.Equal(t, map[string]string{"Bug": ":bug:", "Story": ":book:"}, cfg.Jira.Fields[0].Icons)

	assert.Equal(t, map[string]util.Reaction{"JohnDoe": "star"}, cfg.PullRequest.CustomApproveReaction)
	assert.Equal(t, util.Reaction("fire"), cfg.PullRequest.JiraPriorityReactions["Highest"])

	pluginCfg := struct {
		Mapping map[string]string
	}{}
	require.NoError(t, cfg.LoadCustom("my_plugin", &pluginCfg))
	assert.Equal(t, map[string]string{"FooBar": "baz"}, pluginCfg.Mapping)
}

// see https://github.com/innogames/slack-bot/issues/418
func TestLoadJobsWithoutParameters(t *testing.T) {
	file := writeConfigFile(t, t.TempDir(), "config.yaml", `
jenkins:
  host: https://jenkins.example.com
  jobs:
    JOB1:
    JOB2: {}
    JOB3:
      trigger: foo
`)

	cfg, err := Load(file)
	require.NoError(t, err)

	assert.Equal(t, []string{"JOB1", "JOB2", "JOB3"}, cfg.Jenkins.Jobs.GetSortedNames())
	assert.Equal(t, JobConfig{}, cfg.Jenkins.Jobs["JOB1"])
	assert.Equal(t, JobConfig{}, cfg.Jenkins.Jobs["JOB2"])
	assert.Equal(t, "foo", cfg.Jenkins.Jobs["JOB3"].Trigger)
}

func TestLoadMergesDirectory(t *testing.T) {
	dir := t.TempDir()
	writeConfigFile(t, dir, "a.yaml", `
allowed_users:
  - U1
  - U2
jenkins:
  host: https://jenkins.example.com
  username: bot
  jobs:
    BuildApp:
      trigger: build app
`)
	writeConfigFile(t, dir, "b.yaml", `
allowed_users:
  - U3
jenkins:
  host: https://jenkins2.example.com
  jobs:
    DeployApp:
`)

	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.Equal(t, UserList{"U3"}, cfg.AllowedUsers)
	assert.Equal(t, "https://jenkins2.example.com", cfg.Jenkins.Host)
	assert.Equal(t, "bot", cfg.Jenkins.Username)
	assert.Equal(t, []string{"BuildApp", "DeployApp"}, cfg.Jenkins.Jobs.GetSortedNames())
	assert.Equal(t, "build app", cfg.Jenkins.Jobs["BuildApp"].Trigger)
}

func TestLoadEnvironmentOverrides(t *testing.T) {
	file := writeConfigFile(t, t.TempDir(), "config.yaml", `
allowed_users:
  - U9
jenkins:
  host: https://jenkins.example.com
  username: bot
  approval_timeout: 1m
  jobs:
    BuildApp:
      trigger: build app
`)

	t.Setenv("BOT_JENKINS_HOST", "https://jenkins.env.example.com")
	t.Setenv("BOT_JENKINS_USERNAME", "")
	t.Setenv("BOT_JENKINS_APPROVAL_TIMEOUT", "2m")
	t.Setenv("BOT_JENKINS_JOBS_BUILDAPP_TRIGGER", "build it")
	t.Setenv("BOT_ALLOWED_USERS", "U1,U2")

	cfg, err := Load(file)
	require.NoError(t, err)

	assert.Equal(t, "https://jenkins.env.example.com", cfg.Jenkins.Host)
	assert.Empty(t, cfg.Jenkins.Username)
	assert.Equal(t, "2m0s", cfg.Jenkins.ApprovalTimeout.String())
	assert.Equal(t, "build it", cfg.Jenkins.Jobs["BuildApp"].Trigger)
	assert.Equal(t, UserList{"U1", "U2"}, cfg.AllowedUsers)
}

func TestSetAndLoadCustom(t *testing.T) {
	type customConfig struct {
		Enabled bool
		Name    string
	}

	cfg := Config{}

	// no custom config at all
	loaded := customConfig{}
	require.NoError(t, cfg.LoadCustom("custom", &loaded))
	assert.Equal(t, customConfig{}, loaded)

	// dotted key
	cfg.Set("custom.enabled", true)
	require.NoError(t, cfg.LoadCustom("custom", &loaded))
	assert.Equal(t, customConfig{Enabled: true}, loaded)

	// struct value
	cfg.Set("other", customConfig{Name: "FooBar"})
	other := customConfig{}
	require.NoError(t, cfg.LoadCustom("other", &other))
	assert.Equal(t, customConfig{Name: "FooBar"}, other)

	// unknown key
	unknown := customConfig{}
	require.NoError(t, cfg.LoadCustom("unknown", &unknown))
	assert.Equal(t, customConfig{}, unknown)
}
