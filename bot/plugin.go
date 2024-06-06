package bot

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"github.com/innogames/slack-bot/v2/client"
	log "github.com/sirupsen/logrus"
	"net/rpc"
	"os"
	"os/exec"
)

type SlackBotPluginI interface {
	GetCommands(cfg *Bot, slack client.SlackClient) Commands
}
type SlackBotPlugin struct {
	GetCommands func(cfg *Bot, slack client.SlackClient) Commands
}

func LoadPlugins(b *Bot) Commands {
	commands := Commands{}

	for _, pluginPath := range b.config.Plugins {
		log.Infof("Load plugin %s...", pluginPath)

		c, err := loadPlugin(pluginPath)
		fmt.Println(err)
		c.GetCommands()
	}

	return commands
}

func loadPlugin(pluginPath string) (*CommandRPC, error) {
	// todo use only one logger, and use logrus
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "COMMAND_PLUGIN",
			MagicCookieValue: "bar",
		},
		Plugins: map[string]plugin.Plugin{
			"command": &SlackBotPlugin{},
		},
		Cmd:    exec.Command(pluginPath),
		Logger: logger,
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense("command")
	if err != nil {
		return nil, err
	}

	return raw.(*CommandRPC), nil
}

// ServePlugin serves the plugin
func ServePlugin(impl *SlackBotPlugin) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	fmt.Println("start serving plugin")
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "COMMAND_PLUGIN",
			MagicCookieValue: "bar",
		},
		Plugins: map[string]plugin.Plugin{
			"command": impl,
		},
		//GRPCServer: plugin.DefaultGRPCServer,
		Logger: logger,
	})
	fmt.Println("end serving plugin")
}

// Plugin Server Implementation
func (p *SlackBotPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &CommandRPCServer{Impl: p}, nil
	// todo? return &GreeterRPCServer{Impl: p.Impl, b: b}, nil
}

// Plugin Client Implementation
func (p *SlackBotPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return &CommandRPC{client: c}, nil
}

// Concrete implementation of the plugin Command that wraps RPC client.
type CommandRPC struct {
	client *rpc.Client
}

func (g *CommandRPC) GetCommands() {
	var resp string
	err := g.client.Call("Plugin.GetCommands", new(any), &resp)
	fmt.Println(err)
	fmt.Println(resp)
	if err != nil {
		panic(err)
	}
}

type CommandRPCServer struct {
	// This is the real implementation
	Impl *SlackBotPlugin
}

func (s *CommandRPCServer) GetCommands(args interface{}, resp *string) error {
	commands := s.Impl.GetCommands(&Bot{}, &client.Slack{})
	fmt.Println(commands)
	return nil
}
