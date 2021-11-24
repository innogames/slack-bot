package command

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
)

// NewReplyCommand is a command to reply a message in current thread/channel
func NewListCacheCommand(base bot.BaseCommand) bot.Command {
	return &listCachecommand{base}
}

type listCachecommand struct {
	bot.BaseCommand
}

type distribution struct {
	Id   string
	Name string
}

func (c *listCachecommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("caches", c.listCache),
		matcher.NewPrefixMatcher("clear cache", c.clearCache),
	)
}

func (c *listCachecommand) listCache(match matcher.Result, message msg.Message) {
	dists := []distribution{}
	dists = append(dists, distribution{Id: "E2YF41IFAE1JFI", Name: "klip-dev"})
	dists = append(dists, distribution{Id: "E2YOAZP0AV74Z0", Name: "klip-qa"})

	// show list
	blocks := []slack.Block{}
	for _, v := range dists {
		blocks = append(blocks, slack.NewActionBlock("", client.GetInteractionButton(v.Name, fmt.Sprintf("clear cache %s", v.Id))))
	}
	fmt.Println(blocks)
	c.SendBlockMessage(message, blocks)
}

func (c *listCachecommand) clearCache(match matcher.Result, message msg.Message) {
	text := match.GetString(util.FullMatch)
	sess, err := session.NewSession()
	if nil != err {
		fmt.Println(err.Error())
		return
	}
	svc := cloudfront.New(sess)
	cr := strconv.FormatInt(time.Now().Unix(), 10)
	items := []*string{}
	path := "/kaikas/token-list.json"
	items = append(items, &path)
	size := int64(len(items))
	invalidation := &cloudfront.CreateInvalidationInput{
		DistributionId: &text,
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: &cr,
			Paths: &cloudfront.Paths{
				Items:    items,
				Quantity: &size,
			},
		},
	}

	output, err := svc.CreateInvalidation(invalidation)
	fmt.Println(output)
	fmt.Println("#######")
	fmt.Println(err)
	if nil != err {
		fmt.Println(output)

	} else {
		fmt.Println(err)
	}
	c.SendMessage(message, fmt.Sprintf("cache %s cleared", text))

}
func (c *listCachecommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "list cache <text>",
			Description: "just add the given message",
			Category:    helperCategory,
			Examples: []string{
				"reply Hello, how are you?",
			},
		},
	}
}
