package jenkins

import (
	"net/url"

	"github.com/innogames/slack-bot/v2/bot/config"
)

// resolveJobName resolves a command argument to a configured job. Prefer the
// exact name so encoded slashes in multibranch branch names stay intact. The
// decoded fallback keeps commands using the previous double-encoded form working.
func resolveJobName(jobName string, jobs config.JenkinsJobs) string {
	if _, ok := jobs[jobName]; ok {
		return jobName
	}

	decodedJobName, err := url.PathUnescape(jobName)
	if err == nil {
		if _, ok := jobs[decodedJobName]; ok {
			return decodedJobName
		}
	}

	return jobName
}
