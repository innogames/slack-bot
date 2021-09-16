package jenkins

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStartBuild(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	cfg := config.JobConfig{}
	message := msg.Message{}

	t.Run("error fetching job", func(t *testing.T) {
		client := &mocks.Client{}
		jobName := "TestJob"
		params := Parameters{}

		ctx := context.TODO()
		mocks.AssertReaction(slackClient, iconPending, message)
		client.On("GetJob", ctx, jobName).Return(nil, errors.New("404"))

		err := TriggerJenkinsJob(cfg, jobName, params, slackClient, client, message)

		assert.EqualError(t, err, "Job *TestJob* could not start build with parameters: -none-: 404")
	})

	t.Run("start job", func(t *testing.T) {
		server := spawnJenkinsServer()
		defer server.Close()

		cfg := config.Jenkins{
			Host: server.URL,
		}

		jobCfg := config.JobConfig{}
		client, err := GetClient(cfg)
		assert.Nil(t, err)

		jobName := "testJob"
		params := Parameters{
			"foo": "bar",
		}

		mocks.AssertReaction(slackClient, iconPending, message)
		mocks.AssertRemoveReaction(slackClient, iconPending, message)
		slackClient.On("SendMessage", message, "", mock.Anything).Once().Return("")

		err = TriggerJenkinsJob(jobCfg, jobName, params, slackClient, client, message)
		time.Sleep(time.Millisecond * 100)
		assert.Nil(t, err)
	})

	t.Run("format finish build", func(t *testing.T) {
		build := &gojenkins.Build{}
		build.Raw = &gojenkins.BuildResponse{}
		build.Raw.Duration = float64(60 * 3 * 1000)
		build.Raw.Result = "FAILURE"
		build.Raw.Number = 1233

		build.Raw.URL = "https://jenkins.example.com/job/test/12/"

		actual := getFinishBuildText(build, "user123", "testJob")
		expected := "<@user123> *FAILURE:* testJob #1233 took 3m0s: <https://jenkins.example.com/job/test/12/|Build> <https://jenkins.example.com/job/test/12/console|Console>\nRetry the build by using `retry build testJob #1233`"

		assert.Equal(t, expected, actual)
	})
}

func TestGetAttachment(t *testing.T) {
	t.Run("Success Job", func(t *testing.T) {
		jenkinsBuild := &gojenkins.Build{
			Raw: &gojenkins.BuildResponse{
				Result: gojenkins.STATUS_SUCCESS,
				URL:    "https://jenkins.example.com/build/",
			},
		}
		message := "myMessage"
		actual := getAttachment(jenkinsBuild, message)
		jsonResponse, _ := json.Marshal(actual)

		expected := `{"color":"#00EE00","title":"myMessage","title_link":"https://jenkins.example.com/build/","actions":[{"name":"","text":"Build :white_check_mark:","style":"default","type":"button","url":"https://jenkins.example.com/build/"},{"name":"","text":"Console :page_with_curl:","style":"default","type":"button","url":"https://jenkins.example.com/build/console"},{"name":"","text":"Rebuild :arrows_counterclockwise:","style":"default","type":"button","url":"https://jenkins.example.com/build/rebuild/parameterized"}],"blocks":null}`
		assert.Equal(t, expected, string(jsonResponse))
	})

	t.Run("Running Job", func(t *testing.T) {
		jenkinsBuild := &gojenkins.Build{
			Raw: &gojenkins.BuildResponse{
				Building: true,
				URL:      "https://jenkins.example.com/build/",
			},
		}
		message := "myMessage"
		actual := getAttachment(jenkinsBuild, message)
		jsonResponse, _ := json.Marshal(actual)

		expected := `{"color":"#E0E000","title":"myMessage","title_link":"https://jenkins.example.com/build/","actions":[{"name":"","text":"Build :arrows_counterclockwise:","style":"default","type":"button","url":"https://jenkins.example.com/build/"},{"name":"","text":"Console :page_with_curl:","style":"default","type":"button","url":"https://jenkins.example.com/build/console"},{"name":"","text":"Abort :bomb:","style":"danger","type":"button","url":"https://jenkins.example.com/build/stop/"}],"blocks":null}`
		assert.Equal(t, expected, string(jsonResponse))
	})

	t.Run("Aborted Job", func(t *testing.T) {
		jenkinsBuild := &gojenkins.Build{
			Raw: &gojenkins.BuildResponse{
				Building: false,
				Result:   gojenkins.STATUS_ABORTED,
				URL:      "https://jenkins.example.com/build/",
			},
		}
		message := "myMessage"
		actual := getAttachment(jenkinsBuild, message)
		jsonResponse, _ := json.Marshal(actual)

		expected := `{"color":"#CCCCCC","title":"myMessage","title_link":"https://jenkins.example.com/build/","actions":[{"name":"","text":"Build :black_circle_for_record:","style":"default","type":"button","url":"https://jenkins.example.com/build/"},{"name":"","text":"Console :page_with_curl:","style":"default","type":"button","url":"https://jenkins.example.com/build/console"},{"name":"","text":"Rebuild :arrows_counterclockwise:","style":"default","type":"button","url":"https://jenkins.example.com/build/rebuild/parameterized"}],"blocks":null}`
		assert.Equal(t, expected, string(jsonResponse))
	})

	t.Run("Failed Job", func(t *testing.T) {
		jenkinsBuild := &gojenkins.Build{
			Raw: &gojenkins.BuildResponse{
				Building: false,
				Result:   gojenkins.STATUS_FAIL,
				URL:      "https://jenkins.example.com/build/",
			},
		}
		message := "myMessage"
		actual := getAttachment(jenkinsBuild, message)
		jsonResponse, _ := json.Marshal(actual)

		expected := `{"color":"#CC0000","title":"myMessage","title_link":"https://jenkins.example.com/build/","actions":[{"name":"","text":"Build :x:","style":"default","type":"button","url":"https://jenkins.example.com/build/"},{"name":"","text":"Console :page_with_curl:","style":"default","type":"button","url":"https://jenkins.example.com/build/console"},{"name":"","text":"Rebuild :arrows_counterclockwise:","style":"default","type":"button","url":"https://jenkins.example.com/build/rebuild/parameterized"}],"blocks":null}`
		assert.Equal(t, expected, string(jsonResponse))
	})
}

func TestSendBuildStartedMessage(t *testing.T) {
	slackClient := &mocks.SlackClient{}
	ref := msg.MessageRef{}

	job := &gojenkins.Job{
		Raw: &gojenkins.JobResponse{
			Name: "MyJob",
		},
	}
	build := &gojenkins.Build{
		Raw: &gojenkins.BuildResponse{
			Building: true,
			URL:      "https://jenkins.example.com/build/",
		},
		Job: job,
	}

	mocks.AssertSlackJSON(t, slackClient, ref, `[{"color":"#E0E000","title":"Job MyJob started (#0 - estimated: 0s)","title_link":"https://jenkins.example.com/build/","actions":[{"name":"","text":"Build :arrows_counterclockwise:","style":"default","type":"button","url":"https://jenkins.example.com/build/"},{"name":"","text":"Console :page_with_curl:","style":"default","type":"button","url":"https://jenkins.example.com/build/console"},{"name":"","text":"Abort :bomb:","style":"danger","type":"button","url":"https://jenkins.example.com/build/stop/"}],"blocks":null}]`)

	msgTimestamp := sendBuildStartedMessage(build, slackClient, ref)
	assert.Equal(t, "", msgTimestamp)
}

func spawnJenkinsServer() *httptest.Server {
	mux := http.NewServeMux()

	buildNumber := 42

	// test connection
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`ok`))
	})

	mux.HandleFunc("/job/testJob/api/json", func(w http.ResponseWriter, r *http.Request) {
		job := gojenkins.JobResponse{}
		job.Name = "test"
		job.LastBuild = gojenkins.JobBuild{
			Number: int64(buildNumber),
		}
		encoder := json.NewEncoder(w)
		encoder.Encode(job)
	})

	mux.HandleFunc("/job/notExistingJob/api/json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	mux.HandleFunc("/job/testJob/42/api/json", func(w http.ResponseWriter, r *http.Request) {
		build := gojenkins.BuildResponse{}
		build.Number = 42
		build.Building = true

		encoder := json.NewEncoder(w)
		encoder.Encode(build)
	})
	mux.HandleFunc("/job/testJob/build", func(w http.ResponseWriter, r *http.Request) {
		buildNumber++
		w.Header().Set("Location", "http://foo.bar/job/testJob/111")
	})

	return httptest.NewServer(mux)
}
