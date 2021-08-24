package command

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/slack-go/slack"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/innogames/slack-bot/v2/command/queue"
)

// NewDelayCommand delays the command execution by the given time
func NewDelayCommand(base bot.BaseCommand) bot.Command {
	return &delayCommand{BaseCommand: base, timers: make([]*time.Timer, 0)}
}

type delayCommand struct {
	bot.BaseCommand
	timers []*time.Timer
	mu     sync.Mutex
}

func (c *delayCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher("delay (?P<delay>[\\w]+) (?P<quiet>quiet )?(?P<command>.*)", c.Delay),
		matcher.NewRegexpMatcher("stop (delay|timer) (?P<timer>\\d+)", c.Stop),
	)
}

func (c *delayCommand) Delay(match matcher.Result, message msg.Message) {
	delay, err := util.ParseDuration(match.GetString("delay"))
	if err != nil {
		c.SendMessage(message, "Invalid duration: "+err.Error())
		return
	}

	quietMode := match.GetString("quiet") != ""
	command := match.GetString("command")

	c.mu.Lock()
	defer c.mu.Unlock()

	timer := time.NewTimer(delay)
	c.timers = append(c.timers, timer)

	if !quietMode {
		stopNumber := len(c.timers) - 1
		text := fmt.Sprintf(
			"I queued the command `%s` for %s. Use `stop timer %d` to stop the timer",
			command,
			delay,
			stopNumber,
		)
		blocks := []slack.Block{
			client.GetTextBlock(text),
		}

		// add a abort button, ich we can handle the,
		if c.CanHandleInteractions() {
			blocks = append(
				blocks,
				slack.NewActionBlock(
					"",
					client.GetInteractionButton("Stop timer!", fmt.Sprintf("stop timer %d", stopNumber)),
				),
			)
		}

		c.SendBlockMessage(message, blocks)
	}

	runningCommand := queue.AddRunningCommand(message, "")

	go func() {
		<-timer.C // todo abort here when it was aborted + more random stop key
		runningCommand.Done()

		client.HandleMessage(message.WithText(command))
	}()
}

func (c *delayCommand) Stop(match matcher.Result, message msg.Message) {
	// avoid racing conditions when it's used multiple times in parallel
	c.mu.Lock()
	defer c.mu.Unlock()

	timerNr := match.GetInt("timer")
	if timerNr < len(c.timers) && c.timers[timerNr] != nil {
		c.timers[timerNr].Stop()
		c.timers[timerNr] = nil
		c.SendMessage(message, "Stopped timer!")
	} else {
		c.ReplyError(message, errors.New("invalid timer"))
	}
}

var delayCategory = bot.Category{
	Name:        "Delay",
	Description: "Delay any command by the given time",
	HelpURL:     "https://github.com/innogames/slack-bot#delay",
}

func (c *delayCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "delay <duration> <command>",
			Description: "delay a command by the given offset",
			Examples: []string{
				"delay 1h rely remind me to go to toilet",
				"delay 15m30s trigger job DeployBeta",
				"delay 15min trigger job DeployBeta",
			},
			Category: delayCategory,
		},
		{
			Command:     "stop timer <timerId>",
			Description: "cancel a planned delayCommand",
			Examples: []string{
				"stop timer 1243",
			},
			Category: delayCategory,
		},
	}
}
