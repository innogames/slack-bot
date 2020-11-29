package command

import (
	"errors"
	"fmt"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/bot/msg"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/command/queue"
	"github.com/slack-go/slack"
	"sync"
	"time"
)

// NewDelayCommand delays the command execution by the given time
func NewDelayCommand(slackClient client.SlackClient) bot.Command {
	return &delayCommand{slackClient: slackClient, timers: make([]*time.Timer, 0)}
}

type delayCommand struct {
	slackClient client.SlackClient
	timers      []*time.Timer
	mu          sync.Mutex
}

func (c *delayCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher("delay (?P<delay>[\\w]+) (?P<quiet>quiet )?(?P<command>.*)", c.Delay),
		matcher.NewRegexpMatcher("stop (delay|timer) (?P<timer>\\d+)", c.Stop),
	)
}

func (c *delayCommand) Delay(match matcher.Result, event slack.MessageEvent) {
	delay, err := util.ParseDuration(match.GetString("delay"))
	if err != nil {
		c.slackClient.Reply(event, "Invalid duration: "+err.Error())
		return
	}

	quietMode := match.GetString("quiet") != ""
	command := match.GetString("command")

	timer := time.NewTimer(delay)
	c.timers = append(c.timers, timer)

	if !quietMode {
		c.slackClient.Reply(event, fmt.Sprintf(
			"I queued the command `%s` for %s. Use `stop timer %d` to stop the timer",
			command,
			delay,
			len(c.timers)-1,
		))
	}

	done := queue.AddRunningCommand(event, "")

	go func() {
		<-timer.C // todo abort here when it was aborted + more random stop key
		done <- true

		newMessage := event
		newMessage.Text = command
		client.InternalMessages <- msg.FromSlackEvent(newMessage)
	}()
}

func (c *delayCommand) Stop(match matcher.Result, event slack.MessageEvent) {
	// avoid racing conditions when it's used multiple times in parallel
	c.mu.Lock()
	defer c.mu.Unlock()

	timerNr := match.GetInt("timer")
	if timerNr < len(c.timers) && c.timers[timerNr] != nil {
		c.timers[timerNr].Stop()
		c.timers[timerNr] = nil
		c.slackClient.Reply(event, "Stopped timer!")
	} else {
		c.slackClient.ReplyError(event, errors.New("invalid timer"))
	}
}

var delayCategory = bot.Category{
	Name:    "Delay",
	HelpURL: "https://github.com/innogames/slack-bot#delay",
}

func (c *delayCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "delay",
			Description: "delay a command by the given offset",
			Examples: []string{
				"delay 1h rely remind me to go to toilet",
				"delay 15m30s trigger job DeployBeta",
				"delay 15min trigger job DeployBeta",
			},
			Category: delayCategory,
		},
		{
			Command:     "stop delay",
			Description: "cancel a planned delayCommand",
			Examples: []string{
				"stop timer 1243",
			},
			Category: delayCategory,
		},
	}
}
