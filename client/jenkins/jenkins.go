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

	// we are polling the job every 3-60seconds with a increasing delay
	minDelay      = time.Second * 3
	maxDelay      = time.Minute
	delayIncrease = time.Second * 1
)

type jobResult struct {
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
func WatchBuild(build *gojenkins.Build) <-chan jobResult {
	resultChan := make(chan jobResult, 1)

	go func() {
		defer close(resultChan)

		i := 0
		delay := minDelay
		for {
			time.Sleep(util.GetIncreasingTime(delay, i))
			i++
			if delay <= maxDelay {
				delay += delayIncrease
			}

			build.Poll()
			if !build.IsRunning() {
				resultChan <- jobResult{
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
