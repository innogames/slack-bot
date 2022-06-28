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
	MsgBlocks := []slack.Block{}

	headerObj := slack.NewTextBlockObject("plain_text", "Cloud resource managing functions", false, false)
	if err := headerObj.Validate(); err != nil {
		c.ReplyError(message, err)
		return
	}
	MsgBlocks = append(MsgBlocks, slack.NewHeaderBlock(headerObj))

	MsgBlocks = append(MsgBlocks, slack.NewDividerBlock())

	for _, v := range c.cfg {
		var name string = v.Name
		if v.Alias != "" {
			name = v.Alias
		}
		var description = "no defined description"
		if v.Description != "" {
			description = v.Description
		}
		lambdaDescObj := slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("• %s -> _%s_ \n", name, description), false, false)
		if err := lambdaDescObj.Validate(); err != nil {
			c.ReplyError(message, err)
			return
		}
		MsgBlocks = append(MsgBlocks, slack.NewSectionBlock(lambdaDescObj, nil, nil))
	}

	c.SendBlockMessage(message, MsgBlocks)
}

func (c *lambdaCommand) invoke(match matcher.Result, message msg.Message) {
	target := match.GetString("LAMBDA")
	params := match.GetString("PARAMS")

	lambdaConf := getLambdaConfig(target, c.cfg)
	if lambdaConf == nil {
		c.ReplyError(message, errors.New("no match lambda function"))
		return
	}
	paramsList := strings.Fields(params)
	payload := "{"
	if len(lambdaConf.Inputs) > 0 {
		for i, v := range lambdaConf.Inputs {
			item := ""
			if len(paramsList) > i {
				item = fmt.Sprintf("\"%s\":\"%s\"", v, paramsList[i])
			} else {
				item = fmt.Sprintf("\"%s\":\"\"", v)
			}
			if i < len(lambdaConf.Inputs)-1 {
				item += ","
			}
			payload += item
		}
	}
	payload += "}"
	response, err := c.service.Invoke(&lambda.InvokeInput{
		FunctionName: &lambdaConf.Name,
		Payload:      []byte(payload),
	})
	if err != nil {
		c.ReplyError(message, err)
		return
	}
	unquote, err := strconv.Unquote(string(response.Payload))
	if err != nil {
		c.ReplyError(message, err)
		return
	}
	var output = &config.LambdaOutput{}
	err = json.Unmarshal([]byte(unquote), output)
	if nil != err {
		c.ReplyError(message, err)
		return
	}

	switch output.Code {
	case "500":
		if output.Error != "" {
			c.ReplyError(message, errors.New(output.Error))
		} else {
			c.ReplyError(message, errors.New("something went wrong"))
		}

	case "200":
		blocks := []slack.Block{}
		for _, v := range output.Message {
			for _, key := range lambdaConf.Outputs {
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

func getLambdaConfig(target string, lambdas []config.Lambda) *config.Lambda {
	var conf = &config.Lambda{}
	for _, v := range lambdas {
		if target == v.Name || target == v.Alias {
			conf = &v
			break
		}
	}
	return conf
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
