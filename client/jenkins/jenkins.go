package jenkins

import (
	"fmt"
	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack"
	"net/http"
	"time"
)

const (
	IconRunning = "arrows_counterclockwise"
	IconSuccess = "white_check_mark"
	IconFailed  = "x"
	iconPending = "coffee"
	iconAborted = "black_circle_for_record"
)

type JobResult struct {
	build  *gojenkins.Build
	status string
	output string
}

func GetClient(cfg config.Jenkins) (*gojenkins.Jenkins, error) {
	if !cfg.IsEnabled() {
		return nil, nil
	}

	client := gojenkins.CreateJenkins(
		&http.Client{},
		cfg.Host,
		cfg.Username,
		cfg.Password,
	)

	return client.Init()
}

type Client interface {
	GetJob(id string, parentIDs ...string) (*gojenkins.Job, error)
	BuildJob(name string, options ...interface{}) (int64, error)
	GetQueue() (*gojenkins.Queue, error)
	GetAllNodes() ([]*gojenkins.Node, error)
}

type Job interface {
	Poll() (int, error)
	GetLastBuild() (*gojenkins.Build, error)
	GetBuild(id int64) (*gojenkins.Build, error)
}

// BlockUntilDone will wait until the given build finished (independent from result)
func BlockUntilDone(build *gojenkins.Build) {
	<-WatchBuild(build)
}

// WatchBuild will return a chan which is filled/closed when the build finished
func WatchBuild(build *gojenkins.Build) <-chan JobResult {
	resultChan := make(chan JobResult, 1)

	go func() {
		defer close(resultChan)

		delay := time.Second * 2
		for {
			time.Sleep(delay)
			delay = delay + time.Second*1

			build.Poll()
			if !build.IsRunning() {
				resultChan <- JobResult{
					status: build.GetResult(),
					build:  build,
					output: build.GetConsoleOutput(),
				}

				return
			}
		}
	}()

	return resultChan
}

func processHooks(commands []string, event slack.MessageEvent, params map[string]string) {
	for _, command := range commands {
		temp, err := util.CompileTemplate(command)
		if err != nil {
			fmt.Println(err)
			continue
		}
		text, _ := util.EvalTemplate(temp, params)

		event.Text = text
		client.InternalMessages <- event
	}
}
