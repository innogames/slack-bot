package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/slack-go/slack"
)

// help category to group all Jenkins command
var category = bot.Category{
	Name:        "Cloud-AWS",
	Description: "Interaction with defined aws lambdas",
	HelpURL:     "https://github.com/innogames/slack-bot",
}

type lambdaCommand struct {
	awsCommand
	service *lambda.Lambda
	cfg     []config.Lambda
}
type LambdaReturnCode struct {
	Code string `json:"code"`
}
type LambdaSuccessMessage struct {
	Message []map[string]interface{} `json:"message"`
}

type LambdaFailedMessage struct {
	Message string `json:"message"`
}

// NewAwsCommand is a command to interact with aws resources
func newLambdaCommands(cfg []config.Lambda, base awsCommand) bot.Command {
	svc := lambda.New(base.session)
	return &lambdaCommand{base, svc, cfg}
}

func (c *lambdaCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher("aws show", c.showLambdas),
		matcher.NewRegexpMatcher(`aws run (?P<LAMBDA>[\w|\d|\-|\.|\_]+)(?P<SEP>\s*)(?P<PARAMS>[\w|\d|\-|\.|\_|\s|\,|\S]*)`, c.invoke),
	)
}

func (c *lambdaCommand) showLambdas(match matcher.Result, message msg.Message) {
	blocks := []slack.Block{}
	for _, v := range c.cfg {
		var name string = v.Name
		if v.Alias != "" {
			name = v.Alias
		}
		var description = "no defined description"
		if v.Description != "" {
			description = v.Description
		}
		txtObj := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("\"%s\": %s\n", name, description), false, false)
		if err := txtObj.Validate(); err != nil {
			fmt.Println(err.Error())
			return
		}
		blocks = append(blocks, slack.NewSectionBlock(txtObj, nil, nil))
	}
	c.SendBlockMessage(message, blocks)
}

func (c *lambdaCommand) invoke(match matcher.Result, message msg.Message) {
	target := match.GetString("LAMBDA")
	params := match.GetString("PARAMS")

	lambdaConfig := getLambdaConfig(target, c.cfg)
	if lambdaConfig == nil {
		c.ReplyError(message, errors.New("no match lambda function"))
		return
	}
	// params-> distribution:E2YF41IFAE1JFI paths:/path1,/path2
	var invokePayload []byte
	if params != "" {
		invokePayload = refineParams(params)
	}
	lambdaOutput, err := c.service.Invoke(&lambda.InvokeInput{
		FunctionName: &lambdaConfig.Name,
		Payload:      []byte(invokePayload),
	})
	if err != nil {
		c.ReplyError(message, err)
		return
	}
	unquotedOutput, err := strconv.Unquote(string(lambdaOutput.Payload))
	if err != nil {
		c.ReplyError(message, err)
		return
	}
	var lambdaResponse = &config.LambdaReturnCode{}
	err = json.Unmarshal([]byte(unquotedOutput), lambdaResponse)
	if nil != err {
		c.ReplyError(message, err)
		return
	}

	switch lambdaResponse.Code {
	case "500":
		var resp struct {
			config.LambdaReturnCode
			config.LambdaFailedMessage
		}
		if err := json.Unmarshal([]byte(unquotedOutput), &resp); err != nil {
			c.ReplyError(message, err)
			return
		}
	case "200":
		var resp struct {
			config.LambdaReturnCode
			config.LambdaSuccessMessage
		}
		if err := json.Unmarshal([]byte(unquotedOutput), &resp); err != nil {
			c.ReplyError(message, err)
			return
		}
		blocks := []slack.Block{}
		for _, v := range resp.Message {
			for _, key := range lambdaConfig.Outputs {
				txtObj := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*\"%s\"*: %s\n", key, v[key]), false, false)
				if err := txtObj.Validate(); err != nil {
					c.ReplyError(message, err)
					return
				}
				blocks = append(blocks, slack.NewSectionBlock(txtObj, nil, nil))
			}
			blocks = append(blocks, slack.NewDividerBlock())
		}
		c.SendBlockMessage(message, blocks)

	}
}

func refineParams(params string) []byte {
	inputs := strings.Split(params, " ")
	invokePayload := "{"
	for k, v := range inputs {
		item := strings.Split(v, ":")
		tmp := fmt.Sprintf("\"%s\":\"%s\"", item[0], item[1])
		if k < len(inputs)-1 {
			tmp += ","
		}
		invokePayload += tmp
	}
	invokePayload += "}"
	return []byte(invokePayload)
}

func getLambdaConfig(target string, lambdas []config.Lambda) *config.Lambda {
	var lambdaConfig = &config.Lambda{}
	for _, v := range lambdas {
		if target == v.Name || target == v.Alias {
			lambdaConfig = &v
			break
		}
	}
	return lambdaConfig
}

func (c *lambdaCommand) GetHelp() []bot.Help {
	examples := []string{
		"aws test-lambda a,b,c",
	}

	help := make([]bot.Help, 0)
	help = append(help, bot.Help{
		Command:     "aws run <lambda-name> <params>",
		Description: "invoke selected AWS lambda with given parameters",
		Examples:    examples,
		Category:    category,
	})

	return help
}
