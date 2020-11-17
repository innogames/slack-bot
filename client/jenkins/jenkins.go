package jenkins

import (
	"fmt"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/slack-go/slack"
)

const (
	IconRunning = "arrows_counterclockwise"
	IconSuccess = "white_check_mark"
	IconFailed  = "x"
	iconPending = "coffee"
	iconAborted = "black_circle_for_record"

	// we are polling the job every 10s-5min with a increasing delay
	minDelay = time.Second * 10
	maxDelay = time.Minute * 5
)

type JobResult struct {
	build  *gojenkins.Build
	status string
}

// Job is a interface of gojenkins.Job
type Job interface {
	Poll() (int, error)
	GetLastBuild() (*gojenkins.Build, error)
	GetBuild(id int64) (*gojenkins.Build, error)
}

// WatchBuild will return a chan which is filled/closed when the build finished
func WatchBuild(build *gojenkins.Build) <-chan JobResult {
	resultChan := make(chan JobResult, 1)

	go func() {
		defer close(resultChan)

		delay := util.GetIncreasingDelay(minDelay, maxDelay)
		for {
			time.Sleep(delay.GetNextDelay())

			if !build.IsRunning() {
				resultChan <- JobResult{
					status: build.GetResult(),
					build:  build,
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
