package mqtt

import "github.com/innogames/slack-bot/config"
import mqtt_poho "github.com/eclipse/paho.mqtt.golang"

// GetMqttClient creates a new mqtt client. start it via: docker run -it --rm -p 1883:1883 -p 9001:9001 toke/mosquitto
func GetMqttClient(cfg config.Mqtt) mqtt_poho.Client {
	opts := mqtt_poho.NewClientOptions()
	opts.AddBroker(cfg.Host)
	opts.SetUsername(cfg.Username)
	opts.SetPassword(cfg.Password)

	client := mqtt_poho.NewClient(opts)

	return client
}
