package jenkins

import (
	"testing"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetClient(t *testing.T) {
	cfg := config.Jenkins{}
	jenkinsClient, err := GetClient(cfg)
	assert.Nil(t, err)
	assert.Nil(t, jenkinsClient)
}

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
			{
				Name: "UPPER",
				Type: "upperCase",
			},
			{
				Name: "LOWER",
				Type: "lowerCase",
			},
		},
	}

	params := &Parameters{}
	err := ParseParameters(jobConfig, "", *params)
	assert.Equal(t, &Parameters{}, params)
	assert.Equal(t, "sorry, you have to pass 4 parameters (NAME, VALUE, UPPER, LOWER)", err.Error())

	params = &Parameters{}
	err = ParseParameters(jobConfig, "test ", *params)
	assert.Equal(t, "sorry, you have to pass 4 parameters (NAME, VALUE, UPPER, LOWER)", err.Error())

	params = &Parameters{}
	err = ParseParameters(jobConfig, `testname testvalue "" ""`, *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{
		"NAME":  "testname",
		"VALUE": "testvalue",
		"UPPER": "",
		"LOWER": "",
	}, params)

	params = &Parameters{}
	err = ParseParameters(jobConfig, "testname \"test value\" uPper lOwer", *params)
	assert.Equal(t, nil, err)
	assert.Equal(t, &Parameters{
		"NAME":  "testname",
		"VALUE": "test value",
		"UPPER": "UPPER",
		"LOWER": "lower",
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

	actual = parseWords(`testname "test value" uPper lOwer`)
	assert.Equal(t, []string{"testname", "test value", "uPper", "lOwer"}, actual)
}

func TestWatch(t *testing.T) {
	build := &gojenkins.Build{}

	resultChan := WatchBuild(build)
	assert.Empty(t, resultChan)
}

func TestHook(t *testing.T) {
	ref := msg.MessageRef{}

	t.Run("With template", func(t *testing.T) {
		commands := []string{
			"reply foo",
			"reply {{.var1}}",
		}
		params := Parameters{
			"var1": "bar",
		}
		processHooks(commands, ref, params)
		mocks.AssertQueuedMessage(t, ref.WithText("reply foo"))
		mocks.AssertQueuedMessage(t, ref.WithText("reply bar"))
	})

	t.Run("With Error", func(t *testing.T) {
		commands := []string{
			"reply {{.var1}",
		}
		params := Parameters{}

		processHooks(commands, ref, params)
		assert.Empty(t, client.InternalMessages)
	})
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
