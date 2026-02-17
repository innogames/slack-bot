package aws

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
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
	service *cloudfront.Client
	cfg     []config.AwsCfDistribution
}

// NewAwsCommand is a command to interact with aws resources
func newCloudFrontCommands(cfg []config.AwsCfDistribution, base awsCommand) bot.Command {
	svc := cloudfront.NewFromConfig(base.cfg)
	return &cloudFrontCommand{base, svc, cfg}
}

func (c *cloudFrontCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("aws cf list", c.distributions),
		matcher.NewRegexpMatcher(`aws cf clean (?P<DIST>[\w|\d]+) at (?P<PATH>[\w|\d|\W]+)`, c.clearCache),
	)
}

func (c *cloudFrontCommand) distributions(_ matcher.Result, message msg.Message) {
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
	ctx := context.Background()
	dist := match.GetString("DIST")
	paths := strings.Split(match.GetString("PATH"), ",")

	// Safe conversion to int32 to avoid potential overflow
	pathCount := len(paths)
	if pathCount > 2147483647 { // max int32 value
		c.ReplyError(message, fmt.Errorf("too many paths specified: %d", pathCount))
		return
	}
	quantity := int32(pathCount)

	ref := strconv.FormatInt(time.Now().Unix(), 10)

	invalidation := &cloudfront.CreateInvalidationInput{
		DistributionId: &dist,
		InvalidationBatch: &types.InvalidationBatch{
			CallerReference: &ref,
			Paths: &types.Paths{
				Items:    paths,
				Quantity: &quantity,
			},
		},
	}

	output, err := c.service.CreateInvalidation(ctx, invalidation)

	if output != nil && output.Invalidation != nil {
		fmt.Printf("Invalidation created: %+v\n", output.Invalidation)
		c.SendMessage(message, fmt.Sprintf("cache %s cleared", dist))
	} else {
		if err != nil {
			fmt.Println(err.Error())
		}
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

	return []bot.Help{
		{
			Command:     "aws cf <sub command>",
			Description: "interact with aws cf resources",
			Examples:    examples,
			Category:    category,
		},
	}
}
