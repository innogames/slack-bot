package jenkins

import (
	"fmt"
	"sync"
	"time"

	"github.com/bndr/gojenkins"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

var mu sync.Mutex

// TriggerJenkinsJob starts a new build with given parameters
// it will return when the job was started successfully
// in the background it will watch the current build state and will update the state in the original slack message
func TriggerJenkinsJob(cfg config.JobConfig, jobName string, jobParams map[string]string, slackClient client.SlackClient, jenkins Client, message msg.Message) error {
	log.Infof("%s started started job %s: %s", message.GetUser(), jobName, jobParams)
	_, jobParams[slackUserParameter] = client.GetUser(message.GetUser())

	processHooks(cfg.OnStart, message, jobParams)
	slackClient.AddReaction(iconPending, message)

	build, err := startJob(jenkins, jobName, jobParams)
	if err != nil {
		return errors.Wrapf(err, "Job *%s* could not start job", jobName)
	}

	slackClient.RemoveReaction(iconPending, message)

	estimatedDuration := time.Duration(build.Raw.EstimatedDuration) * time.Millisecond
	text := fmt.Sprintf(
		"Job %s started (#%d - estimated: %s)",
		build.Job.GetName(),
		build.GetBuildNumber(),
		util.FormatDuration(estimatedDuration),
	)

	// send main response (with parameters)
	msgTimestamp := slackClient.SendMessage(
		message,
		"",
		GetAttachment(build, text),
	)

	runningCommand := queue.AddRunningCommand(
		message,
		fmt.Sprintf("inform job %s #%d", jobName, build.GetBuildNumber()),
	)
	go func() {
		// wait until job is not running anymore
		<-WatchBuild(build)
		runningCommand.Done()

		// update main message
		attachment := GetAttachment(build, fmt.Sprintf(
			"Job %s #%d finished!",
			build.Job.GetName(),
			build.GetBuildNumber(),
		))

		slackClient.SendMessage(
			message,
			"",
			slack.MsgOptionUpdate(msgTimestamp),
			attachment,
		)

		duration := time.Duration(build.GetDuration()) * time.Millisecond
		text = fmt.Sprintf(
			"<@%s> *%s:* %s #%d took %s: <%s|Build> <%sconsole/|Console>",
			message.GetUser(),
			build.GetResult(),
			jobName,
			build.GetBuildNumber(),
			util.FormatDuration(duration),
			build.GetUrl(),
			build.GetUrl(),
		)
		if build.IsGood() {
			slackClient.SendMessage(message, text)
			processHooks(cfg.OnSuccess, message, jobParams)
		} else {
			// failed/aborted build
			text += fmt.Sprintf("\nRetry the build by using `retry build %s #%d`", jobName, build.GetBuildNumber())

			slackClient.SendMessage(
				message,
				text,
				slack.MsgOptionTS(msgTimestamp),
			)
			processHooks(cfg.OnFailure, message, jobParams)
		}
	}()

	return nil
}

// startJob starts a job and waits until job is not queued anymore
func startJob(jenkins Client, jobName string, jobParams map[string]string) (*gojenkins.Build, error) {
	// avoid nasty racing conditions when two people are starting the same job
	mu.Lock()
	defer mu.Unlock()

	job, err := jenkins.GetJob(jobName)
	if err != nil {
		return nil, err
	}

	lastBuildID := job.Raw.LastBuild.Number

	_, err = job.InvokeSimple(jobParams)
	if err != nil {
		return nil, err
	}

	var newBuildID int64

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	// wait until build ios really really running not just queued
	for range ticker.C {
		job.Poll()

		newBuildID = job.Raw.LastBuild.Number
		if newBuildID > lastBuildID {
			break
		}
	}

	log.
		WithField("job", jobName).
		Infof("Queued job %s #%d", jobName, newBuildID)

	return job.GetBuild(newBuildID)
}

// GetAttachment creates a attachment object for a given build
func GetAttachment(build *gojenkins.Build, message string) slack.MsgOption {
	attachment := getAttachment(build, message)

	return slack.MsgOptionAttachments(attachment)
}

func getAttachment(build *gojenkins.Build, message string) slack.Attachment {
	var icon string
	var color string
	if build.Raw.Building {
		icon = IconRunning
		color = "#E0E000"
	} else if build.Raw.Result == gojenkins.STATUS_SUCCESS {
		icon = IconSuccess
		color = "#00EE00"
	} else if build.Raw.Result == gojenkins.STATUS_ABORTED {
		icon = iconAborted
		color = "#CCCCCC"
	} else {
		icon = IconFailed
		color = "#CC0000"
	}

	attachment := slack.Attachment{
		Title:     message,
		TitleLink: build.GetUrl(),
		Color:     color,
	}

	for _, param := range build.GetParameters() {
		if param.Value == "" || param.Name == slackUserParameter {
			continue
		}

		attachment.Fields = append(attachment.Fields, slack.AttachmentField{
			Title: param.Name,
			Value: param.Value,
			Short: true,
		})
	}

	attachment.Actions = []slack.AttachmentAction{
		client.GetSlackLink(fmt.Sprintf("Build :%s:", icon), build.GetUrl()),
		client.GetSlackLink("Console :page_with_curl:", build.GetUrl()+"console"),
	}

	if build.Raw.Building {
		attachment.Actions = append(
			attachment.Actions,
			client.GetSlackLink("Abort :bomb:", build.GetUrl()+"stop/", "danger"),
		)
	} else {
		attachment.Actions = append(
			attachment.Actions,
			client.GetSlackLink("Rebuild :arrows_counterclockwise:", build.GetUrl()+"rebuild/parameterized"),
		)
	}

	return attachment
}
