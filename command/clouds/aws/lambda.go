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
	"github.com/innogames/slack-bot/v2/client"
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
	cfg     config.Aws
	alias   string
}
type LambdaResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewAwsCommand is a command to interact with aws resources
func newLambdaCommands(cfg config.Aws, base awsCommand) bot.Command {
	svc := lambda.New(base.session)
	cmd := "aws"
	if cfg.Alias != "" {
		cmd = cfg.Alias
	}
	return &lambdaCommand{base, svc, cfg, cmd}
}

func (c *lambdaCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewTextMatcher(fmt.Sprintf("%s show", c.alias), c.showLambdas),
		matcher.NewRegexpMatcher(`choose (?P<choose>[\w\s\-]+)`, c.choose),
		matcher.NewRegexpMatcher(`lambda_invoke (?P<invoke>[\w\s\-]+)`, c.invoke),
	)
}

func (c *lambdaCommand) showLambdas(match matcher.Result, message msg.Message) {
	msgBlock := []slack.Block{}

	headerObj := slack.NewTextBlockObject("plain_text", "Cloud resource managing functions", false, false)
	if err := headerObj.Validate(); err != nil {
		c.ReplyError(message, err)
		return
	}
	msgBlock = append(msgBlock, slack.NewHeaderBlock(headerObj))

	msgBlock = append(msgBlock, slack.NewDividerBlock())

	for _, v := range c.cfg.Lambda {
		name := v.Name
		description := "no defined description"
		if v.Desc != "" {
			description = v.Desc
		}
		funcName := v.FuncName

		button := slack.NewActionBlock(
			"",
			client.GetInteractionButton(name, fmt.Sprintf("choose %s", funcName)),
		)
		desc := client.GetTextBlock(description)

		msgBlock = append(msgBlock, button)
		msgBlock = append(msgBlock, desc)
		msgBlock = append(msgBlock, slack.NewDividerBlock())
	}

	c.SendBlockMessage(message, msgBlock)
}

func (c *lambdaCommand) choose(match matcher.Result, message msg.Message) {
	choose := match.GetString("choose")
	submitText := slack.NewTextBlockObject(slack.PlainTextType, "Submit", false, false)
	var modalRequest slack.ModalViewRequest

	for _, v := range c.cfg.Lambda {
		if v.FuncName == choose {
			if len(v.Inputs) != 0 {
				blocks := []slack.Block{}
				for _, val := range v.Inputs {
					block := slack.NewInputBlock(
						fmt.Sprintf("%s-block", val.Key),
						slack.NewTextBlockObject(slack.PlainTextType, val.Key, false, false),
						slack.NewTextBlockObject(slack.PlainTextType, val.Desc, false, false),
						slack.NewPlainTextInputBlockElement(
							slack.NewTextBlockObject(slack.PlainTextType, val.Desc, false, false),
							val.Key),
					)
					blocks = append(blocks, block)
				}
				modalRequest.Type = slack.ViewType("modal")
				modalRequest.Title = slack.NewTextBlockObject(slack.PlainTextType, fmt.Sprintf("Run %s", v.Name), false, false)
				modalRequest.Submit = submitText
				modalRequest.Blocks = slack.Blocks{
					BlockSet: blocks,
				}
				modalRequest.CallbackID = fmt.Sprintf("lambda_invoke %s", choose)
				modalRequest.PrivateMetadata = message.GetMessageRef().Channel

				c.OpenView(message.GetTriggerID(), modalRequest)
			} else {
				resp, err := c.call(&choose, "")
				if err != nil {
					c.ReplyError(message, err)
					return
				}
				switch resp.Code {
				case "200":
					c.SendMessage(message, resp.Message)
				default:
					c.ReplyError(message, fmt.Errorf("failed to run command %s with code %s and error %s", choose, resp.Code, resp.Message))
				}
			}
			break
		}

	}

}

func (c *lambdaCommand) invoke(match matcher.Result, message msg.Message) {
	invoke := match.GetString("invoke")
	modalResp := message.MessageRef.View.State.Values
	partials := []string{}
	if len(modalResp) == 0 {
		c.ReplyError(message, errors.New("empty modal response"))
		return
	}
	for _, outer := range modalResp {
		for k, v := range outer {
			partials = append(partials, fmt.Sprintf("\"%s\":\"%s\"", k, v.Value))
		}
	}
	req := fmt.Sprintf("{%s}", strings.Join(partials, ","))
	resp, err := c.call(&invoke, req)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	switch resp.Code {
	case "200":
		c.SendMessage(message, fmt.Sprintf("Successfully run command %s with body %s", invoke, req))
	default:
		c.ReplyError(message, fmt.Errorf("failed to run command %s with code %s and error %s", invoke, resp.Code, resp.Message))
	}
}

func (c *lambdaCommand) call(funcName *string, request string) (*LambdaResponse, error) {
	lambdaResp, err := c.service.Invoke(&lambda.InvokeInput{
		FunctionName: funcName,
		Payload:      []byte(request),
	})

	if err != nil {
		return nil, err
	}
	unquote, err := strconv.Unquote(string(lambdaResp.Payload))
	if err != nil {
		return nil, err
	}
	resp := &LambdaResponse{}
	err = json.Unmarshal([]byte(unquote), resp)
	if err != nil {
		return nil, err
	}
	return resp, nil

}

func (c *lambdaCommand) GetHelp() []bot.Help {
	examples := []string{
		fmt.Sprintf("%s show", c.alias),
	}
	help := make([]bot.Help, 0)
	help = append(help, bot.Help{
		Command:     fmt.Sprintf("%s show", c.alias),
		Description: "invoke selected AWS lambda with given parameters",
		Examples:    examples,
		Category:    category,
	})

	return help
}
