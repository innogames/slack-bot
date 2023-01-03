package pool

import (
	"fmt"
	"strings"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
)

// newPoolCommands display usage of the pool
func newPoolCommands(slackClient client.SlackClient, cfg *config.Pool, p *pool) bot.Command {
	return &poolCommands{slackClient, cfg, p}
}

type poolCommands struct {
	slackClient client.SlackClient
	config      *config.Pool
	pool        *pool
}

// IsEnabled can be switched on / off via config
func (c *poolCommands) IsEnabled() bool {
	return c.config.IsEnabled()
}

// Matcher to handle commands
func (c *poolCommands) GetMatcher() matcher.Matcher {
	var resources []string
	for _, res := range c.config.Resources {
		resources = append(resources, res.Name)
	}
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(fmt.Sprintf("pool lock\\b( )?(?P<resource>(%s\\b))?(( )?(?P<reason>.+))?", strings.Join(resources, "\\b|")), c.lockResource),
		matcher.NewRegexpMatcher(fmt.Sprintf("pool unlock( )?(?P<resource>(%s))?", strings.Join(resources, "|")), c.unlockResource),
		matcher.NewRegexpMatcher("pool locks", c.listUserResources),
		matcher.NewRegexpMatcher("pool list( )?(?P<status>(free|used|locked))?", c.listResources),
		matcher.NewRegexpMatcher("pool info( )?(?P<status>(free|used|locked))?", c.listPoolInfo),
		matcher.NewRegexpMatcher(fmt.Sprintf("pool extend (?P<resource>(%s)) (?P<duration>([0-9]+[hmsd]))", strings.Join(resources, "|")), c.extend),
	)
}

// RunAsync function to observe, notify and unlock expired locks
func (c *poolCommands) RunAsync() {
	for {
		now := time.Now()
		nowIn := now.Add(c.config.NotifyExpire)
		allLocks := c.pool.GetLocks("")
		for _, lock := range allLocks {
			if now.After(lock.LockUntil) && lock.WarningSend {
				_ = c.pool.Unlock(lock.User, lock.Resource.Name)
				c.slackClient.SendToUser(lock.User, fmt.Sprintf("your lock for `%s` expired and got removed", lock.Resource.Name))
				continue
			}

			if nowIn.After(lock.LockUntil) && !lock.WarningSend {
				blocks := []slack.Block{
					client.GetTextBlock(
						fmt.Sprintf("your lock for `%s` is going to expire at %s.\nextend your lock if you need it longer.", lock.Resource.Name, lock.LockUntil.Format(time.RFC1123)),
					),
					slack.NewActionBlock(
						"extend_30m",
						client.GetInteractionButton("action_30m", "30 mins", fmt.Sprintf("pool extend %s 30m", lock.Resource.Name)),
						client.GetInteractionButton("action_1h", "1 hour", fmt.Sprintf("pool extend %s 1h", lock.Resource.Name)),
						client.GetInteractionButton("action_2h", "2 hours", fmt.Sprintf("pool extend %s 2h", lock.Resource.Name)),
						client.GetInteractionButton("action_1d", "1 day", fmt.Sprintf("pool extend %s 24h", lock.Resource.Name)),
						client.GetInteractionButton("action_unlock", "unlock now!", fmt.Sprintf("pool unlock %s", lock.Resource.Name)),
					),
				}
				c.slackClient.SendBlockMessageToUser(lock.User, blocks)
				lock.WarningSend = true
			}
		}

		time.Sleep(1 * time.Minute)
	}
}

func (c *poolCommands) lockResource(match matcher.Result, message msg.Message) {
	_, userName := client.GetUserIDAndName(message.GetUser())

	resourceName := match.GetString("resource")
	reason := match.GetString("reason")

	resource, err := c.pool.Lock(userName, reason, resourceName)
	if err != nil {
		c.slackClient.ReplyError(message, err)
		return
	}
	c.slackClient.SendMessage(message, fmt.Sprintf("`%s` is locked for you until %s!\n%s%s", resource.Resource.Name, resource.LockUntil.Format(time.RFC1123), getFormattedReason(resource.Reason), getAddressesAndFeatures(&resource.Resource)))
}

func (c *poolCommands) unlockResource(match matcher.Result, message msg.Message) {
	_, userName := client.GetUserIDAndName(message.GetUser())

	resourceName := match.GetString("resource")

	if len(resourceName) == 0 {
		lockedByUser := c.pool.GetLocks(userName)
		if len(lockedByUser) == 0 {
			c.slackClient.SendMessage(message, "you don't have any locks")
			return
		}

		if len(lockedByUser) > 1 {
			var locks []string
			for _, lock := range lockedByUser {
				locks = append(locks, fmt.Sprintf("`%s` until %s\n%s", lock.Resource.Name, lock.LockUntil.Format(time.RFC1123), getFormattedReason(lock.Reason)))
			}
			c.slackClient.SendMessage(message, fmt.Sprintf("which one should be unlocked?\n%s", strings.Join(locks, "\n")))
			return
		}

		resourceName = lockedByUser[0].Resource.Name
	}

	err := c.pool.Unlock(userName, resourceName)
	if err != nil {
		c.slackClient.ReplyError(message, err)
		return
	}
	c.slackClient.SendMessage(message, fmt.Sprintf("`%s` is free again", resourceName))
}

func (c *poolCommands) extend(match matcher.Result, message msg.Message) {
	_, userName := client.GetUserIDAndName(message.GetUser())

	resourceName := match.GetString("resource")
	duration := match.GetString("duration")

	res, err := c.pool.ExtendLock(userName, resourceName, duration)
	if err != nil {
		c.slackClient.ReplyError(message, err)
		return
	}
	if res == nil {
		c.slackClient.ReplyError(message, fmt.Errorf("%s expired already", resourceName))
		return
	}

	c.slackClient.SendMessage(message, fmt.Sprintf("`%s` got extended until %s", resourceName, res.LockUntil.Format(time.RFC1123)))
}

func (c *poolCommands) listResources(match matcher.Result, message msg.Message) {
	status := match.GetString("status")

	var messages []string
	if len(status) == 0 || status == "free" {
		messages = append(messages, "*Available:*")
		free := c.pool.GetFree()
		var resources []string
		for _, f := range free {
			resources = append(resources, fmt.Sprintf("`%s`", f.Name))
		}
		messages = append(messages, strings.Join(resources, ", "))
	}
	messages = append(messages, "")
	if len(status) == 0 || status == "used" || status == "locked" {
		locked := c.pool.GetLocks("")
		messages = append(messages, "*Used/Locked:*")
		for _, l := range locked {
			messages = append(messages, fmt.Sprintf("`%s` locked by %s until %s\n%s", l.Resource.Name, l.User, l.LockUntil.Format(time.RFC1123), getFormattedReason(l.Reason)))
		}
	}

	c.slackClient.SendMessage(message, strings.Join(messages, "\n"))
}

func (c *poolCommands) listUserResources(match matcher.Result, message msg.Message) {
	_, userName := client.GetUserIDAndName(message.GetUser())

	lockedByUser := c.pool.GetLocks(userName)
	if len(lockedByUser) == 0 {
		c.slackClient.SendMessage(message, "you don't have any locks")
		return
	}

	locks := []string{"*Your locks:*\n"}
	for _, lock := range lockedByUser {
		locks = append(locks, fmt.Sprintf("`%s` until %s\n%s", lock.Resource.Name, lock.LockUntil.Format(time.RFC1123), getFormattedReason(lock.Reason)))
	}
	c.slackClient.SendMessage(message, fmt.Sprintf(" %s", strings.Join(locks, "\n")))
}

func (c *poolCommands) listPoolInfo(match matcher.Result, message msg.Message) {
	status := match.GetString("status")

	var messages []string
	if len(status) == 0 || status == "free" {
		messages = append(messages, "*Available:*")
		free := c.pool.GetFree()
		for _, f := range free {
			messages = append(messages, fmt.Sprintf("`%s`:\n%s\n", f.Name, getAddressesAndFeatures(f)))
		}
	}
	messages = append(messages, "")
	if len(status) == 0 || status == "used" || status == "locked" {
		locked := c.pool.GetLocks("")
		messages = append(messages, "*Used/Locked:*")
		for _, l := range locked {
			messages = append(messages, fmt.Sprintf("`%s`:\n locked by %s until %s\n%s%s", l.Resource.Name, l.User, l.LockUntil.Format(time.RFC1123), getFormattedReason(l.Reason), getAddressesAndFeatures(&l.Resource)))
		}
	}

	c.slackClient.SendMessage(message, strings.Join(messages, "\n"))
}

func getFormattedReason(reason string) string {
	if len(reason) == 0 {
		return ""
	}
	return fmt.Sprintf("_%s_\n", reason)
}

func getAddressesAndFeatures(resource *config.Resource) string {
	var lines []string
	lines = append(lines, ">_Addresses:_")
	for _, address := range resource.Addresses {
		lines = append(lines, fmt.Sprintf(">- %s", address))
	}

	lines = append(lines, ">_Features:_")
	for _, address := range resource.Features {
		lines = append(lines, fmt.Sprintf(">- %s", address))
	}
	return strings.Join(lines, "\n")
}

// GetHelp documentation about the command and how to use it
func (c *poolCommands) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "pool list <free, locked>",
			Description: "display available / free resources of the pool",
			Category:    category,
			Examples: []string{
				"pool list _list all resources_",
				"pool list free _list available resources only_",
				"pool list locked _list locked resources only_",
			},
		},
		{
			Command:     "pool info",
			Description: "list detailed infos about the pool",
			Category:    category,
			Examples: []string{
				"pool info _list detailed infos about resources_",
			},
		},
		{
			Command:     "pool lock <resource> <reason>",
			Description: "lock a resource with a reason, resource and reason are optional, no resource will lock any available one",
			Category:    category,
			Examples: []string{
				"pool lock _lock an available resource_",
				"pool lock with reason _lock an available resource with a reason_",
				"pool lock xa _lock a specific resource_",
				"pool lock xa with reason _lock a specific resource with a reason_",
			},
		},
		{
			Command:     "pool locks",
			Description: "show your locked resources",
			Category:    category,
			Examples: []string{
				"pool locks _show your current locks_",
			},
		},
		{
			Command:     "pool unlock <resource>",
			Description: "unlock your locked and/or specific resource, if you have more then one resource locked displays the list of locked resources",
			Category:    category,
			Examples: []string{
				"pool unlock _unlock your resource_",
				"pool unlock xa _lock a specific resource_",
			},
		},
		{
			Command:     "pool extend <resource> <duration>",
			Description: "extend the time a resource is locked",
			Category:    category,
			Examples: []string{
				"pool extend xa 30m _extend lock of resource xa by 30mins_",
			},
		},
	}
}
