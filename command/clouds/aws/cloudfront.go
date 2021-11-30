package aws

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/client"
	"github.com/slack-go/slack"
)

// command to trigger/start jenkins jobs
type cloudFrontCommand struct {
	awsCommand
	service *cloudfront.CloudFront
	cfg     []config.AwsCfDistribution
}

// NewAwsCommand is a command to interact with aws resources
func newCloudFrontCommands(cfg []config.AwsCfDistribution, base awsCommand) bot.Command {
	svc := cloudfront.New(base.session)
	return &cloudFrontCommand{base, svc, cfg}
}

func (c *cloudFrontCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("aws cf list", c.distributions),
		matcher.NewRegexpMatcher(`aws cf clean (?P<DIST>[\w|\d]+) at (?P<PATH>[\w|\d|\W]+)`, c.clearCache),
	)
}

func (c *cloudFrontCommand) distributions(match matcher.Result, message msg.Message) {
	// show list
	blocks := []slack.Block{}
	for _, v := range c.cfg {
		txtObj := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("\"%s\": %s\n", v.ID, v.Name), false, false)
		if err := txtObj.Validate(); err != nil {
			fmt.Println(err.Error())
			return
		}
		blocks = append(blocks, slack.NewSectionBlock(txtObj, nil, nil))
	}
	c.SendBlockMessage(message, blocks)
}

func (c *cloudFrontCommand) clearCache(match matcher.Result, message msg.Message) {
	dist := match.GetString("DIST")
	paths := []*string{}
	for _, v := range strings.Split(match.GetString("PATH"), ",") {
		_tempVal := v
		paths = append(paths, &_tempVal)
	}
	quantity := int64(len(paths))
	ref := strconv.FormatInt(time.Now().Unix(), 10)

	invalidation := &cloudfront.CreateInvalidationInput{
		DistributionId: &dist,
		InvalidationBatch: &cloudfront.InvalidationBatch{
			CallerReference: &ref,
			Paths: &cloudfront.Paths{
				Items:    paths,
				Quantity: &quantity,
			},
		},
	}

	output, err := c.service.CreateInvalidation(invalidation)

	if nil != output.Invalidation {
		fmt.Println(output.String())
		c.SendMessage(message, fmt.Sprintf("cache %s cleared", dist))
	} else {
		fmt.Println(err.Error())
		blocks := []slack.Block{
			client.GetTextBlock(
				"Oops! Command `" + message.GetText() + "` failed...",
			),
		}
		c.SendBlockMessage(message, blocks)

		return
	}
}

func (c *cloudFrontCommand) GetHelp() []bot.Help {
	examples := []string{
		"aws cf clean cache",
	}

	help := make([]bot.Help, 0)
	help = append(help, bot.Help{
		Command:     "aws cf <sub command>",
		Description: "interact with aws cf resources",
		Examples:    examples,
		Category:    category,
	})

	return help
}
