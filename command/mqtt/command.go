package mqtt

import (
	"fmt"
	mqtt_poho "github.com/eclipse/paho.mqtt.golang"
	"github.com/innogames/slack-bot/bot"
	"github.com/innogames/slack-bot/bot/matcher"
	"github.com/innogames/slack-bot/client"
	"github.com/innogames/slack-bot/config"
	"github.com/nlopes/slack"
)

// NewMqttCommand is able to read/write data to a mqtt topic
func NewMqttCommand(slackClient client.SlackClient, cfg config.Mqtt) bot.Command {
	if !cfg.IsEnabled() {
		return nil
	}

	mqttClient := GetMqttClient(cfg)
	token := mqttClient.Connect()
	token.Wait()
	if token.Error() != nil {
		fmt.Println(token.Error())
	}
	return &mqttCommand{
		slackClient,
		mqttClient,
		cfg,
	}
}

type mqttCommand struct {
	slackClient client.SlackClient
	mqtt        mqtt_poho.Client
	cfg         config.Mqtt
}

func (c *mqttCommand) GetMatcher() matcher.Matcher {
	return matcher.NewGroupMatcher(
		matcher.NewRegexpMatcher(`mqtt publish (?P<topic>.+) (?P<value>.+)`, c.Publish),
		matcher.NewRegexpMatcher(`mqtt subscribe (?P<topic>.+)`, c.Subscribe),
		matcher.NewRegexpMatcher(`mqtt unsubscribe (?P<topic>.+)`, c.Unsubscribe),
	)
}

func (c *mqttCommand) Unsubscribe(match matcher.Result, event slack.MessageEvent) {
	topic := match.GetString("topic")

	token := c.mqtt.Unsubscribe(topic)
	token.Wait()

	if token.Error() != nil {
		c.slackClient.ReplyError(event, token.Error())
		return
	}

	c.slackClient.Reply(event, fmt.Sprintf("Subscribed to '%s'", topic))
}

func (c *mqttCommand) Subscribe(match matcher.Result, event slack.MessageEvent) {
	topic := match.GetString("topic")

	token := c.mqtt.Subscribe(topic, 1, func(_ mqtt_poho.Client, message mqtt_poho.Message) {
		c.slackClient.Reply(event, fmt.Sprintf("New message in `%s`: '%s'", topic, message.Payload()))
	})
	token.Wait()
	if token.Error() != nil {
		c.slackClient.ReplyError(event, token.Error())
		return
	}

	c.slackClient.Reply(event, fmt.Sprintf("Subscribed to '%s'", topic))
}

func (c *mqttCommand) Publish(match matcher.Result, event slack.MessageEvent) {
	topic := match.GetString("topic")
	value := match.GetString("value")

	token := c.mqtt.Publish(topic, 1, false, value)
	token.Wait()
	if token.Error() != nil {
		c.slackClient.ReplyError(event, token.Error())
		return
	}

	c.slackClient.Reply(event, fmt.Sprintf("Publish '%s' to '%s'", value, topic))
}

func (c *mqttCommand) GetHelp() []bot.Help {
	return []bot.Help{
		{
			Command:     "mqtt subscribe",
			Description: "Subscribes to a MQTT topic",
			Examples: []string{
				"mqtt subscribe temperature",
			},
		},
		{
			Command:     "mqtt unsubscribe",
			Description: "Unsubscribes to a MQTT topic",
			Examples: []string{
				"mqtt unsubscribe temperature",
			},
		},
		{
			Command:     "mqtt publish",
			Description: "Publish a value to a MQTT topic",
			Examples: []string{
				"mqtt publish temperature 1.00",
			},
		},
	}
}
