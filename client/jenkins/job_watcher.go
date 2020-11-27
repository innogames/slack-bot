package jenkins

import (
	"time"

	"github.com/bndr/gojenkins"
)

const watchInterval = time.Second * 15

// WatchJob returns a chan which gets notified about any (finished) build of the job
func WatchJob(jenkins Client, jobName string, stop chan bool) (chan gojenkins.Build, error) {
	job, err := jenkins.GetJob(jobName)

	if err != nil {
		return nil, err
	}
	lastBuild, err := job.GetLastBuild()
	if err != nil {
		return nil, err
	}

	returnChan := make(chan gojenkins.Build, 1)

	go func() {
		timer := time.NewTicker(watchInterval)
		defer timer.Stop()

		for {
			select {
			case <-stop:
				return
			case <-timer.C:
				job.Poll()

				build, _ := job.GetLastBuild()
				if build == nil || build.Raw.Building {
					continue
				}

				if build.GetBuildNumber() != lastBuild.GetBuildNumber() {
					returnChan <- *build
					lastBuild = build
				}
			}
		}
	}()

	return returnChan, nil
}
