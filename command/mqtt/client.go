package mqtt

import "github.com/innogames/slack-bot/config"
import mqtt_poho "github.com/eclipse/paho.mqtt.golang"

// GetMqttClient creates a new mqtt client
func GetMqttClient(cfg config.Mqtt) mqtt_poho.Client {
	opts := mqtt_poho.NewClientOptions()
	opts.AddBroker(cfg.Host)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)

	client := mqtt_poho.NewClient(opts)

	return client
}
