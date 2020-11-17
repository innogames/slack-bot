package jenkins

import (
	"github.com/innogames/slack-bot/bot/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJenkinsNoParameters(t *testing.T) {
	jobConfig := config.JobConfig{}

	params := &Parameters{}
	err := ParseParameters(jobConfig, "", *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{}, params)

	params = &Parameters{}
	err = ParseParameters(jobConfig, "test", *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{}, params)
}

func TestJenkinsParameters(t *testing.T) {
	jobConfig := config.JobConfig{
		Parameters: []config.JobParameter{
			{Name: "NAME"},
			{Name: "VALUE"},
		},
	}

	params := &Parameters{}
	err := ParseParameters(jobConfig, "", *params)
	assert.Equal(t, &Parameters{}, params)
	assert.Equal(t, "sorry, you have to pass 2 parameters (NAME, VALUE)", err.Error())

	params = &Parameters{}
	err = ParseParameters(jobConfig, "test ", *params)
	assert.Equal(t, "sorry, you have to pass 2 parameters (NAME, VALUE)", err.Error())

	params = &Parameters{}
	err = ParseParameters(jobConfig, "testname testvalue", *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{
		"NAME":  "testname",
		"VALUE": "testvalue",
	}, params)

	params = &Parameters{}
	err = ParseParameters(jobConfig, "testname \"test value\"", *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{
		"NAME":  "testname",
		"VALUE": "test value",
	}, params)
}

func TestJenkinsDefaultParameters(t *testing.T) {
	jobConfig := config.JobConfig{
		Parameters: []config.JobParameter{
			{Name: "NAME"},
			{Name: "FLAG", Type: "bool"},
			{Name: "VALUE", Default: "defaultValue"},
		},
	}

	params := &Parameters{}
	err := ParseParameters(jobConfig, "", *params)
	assert.Equal(t, &Parameters{}, params)
	assert.Equal(t, "sorry, you have to pass 3 parameters (NAME, FLAG, VALUE)", err.Error())

	params = &Parameters{}
	err = ParseParameters(jobConfig, "testname TRUE", *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{
		"NAME":  "testname",
		"FLAG":  "true",
		"VALUE": "defaultValue",
	}, params)

	params = &Parameters{}
	err = ParseParameters(jobConfig, "testname false testvalue", *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{
		"NAME":  "testname",
		"FLAG":  "false",
		"VALUE": "testvalue",
	}, params)
}

func TestParseWords(t *testing.T) {
	actual := parseWords("")
	assert.Equal(t, []string{}, actual)

	actual = parseWords("test test2")
	assert.Equal(t, []string{"test", "test2"}, actual)

	actual = parseWords("test \"test2\"")
	assert.Equal(t, []string{"test", "test2"}, actual)

	actual = parseWords("test \"der test\"")
	assert.Equal(t, []string{"test", "der test"}, actual)

	actual = parseWords("test \"der test")
	assert.Equal(t, []string{"test", "der test"}, actual)
}

func TestJenkinsMixedParameters(t *testing.T) {
	jobConfig := config.JobConfig{
		Parameters: []config.JobParameter{
			{Name: "NAME"},
			{Name: "SUBTYPE"},
			{Name: "VALUE", Default: "defaultValue"},
		},
	}

	params := &Parameters{
		"SUBTYPE": "mySubtype",
	}
	err := ParseParameters(jobConfig, "", *params)
	assert.Equal(t, "sorry, you have to pass 3 parameters (NAME, SUBTYPE, VALUE)", err.Error())

	params = &Parameters{
		"SUBTYPE": "mySubtype",
	}
	err = ParseParameters(jobConfig, "testname", *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{
		"NAME":    "testname",
		"SUBTYPE": "mySubtype",
		"VALUE":   "defaultValue",
	}, params)

	params = &Parameters{}
	err = ParseParameters(jobConfig, "testname testsubtype testvalue", *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{
		"NAME":    "testname",
		"SUBTYPE": "testsubtype",
		"VALUE":   "testvalue",
	}, params)
}
