package bot

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net/rpc"
	"os"
	"os/exec"
)

// Interface for SlackBotPlugin to get commands
type Plugin interface {
	GetCommands() string
}

// RPC Client structure
type SlackBotPluginRPC struct {
	client *rpc.Client
}

func (g *SlackBotPluginRPC) GetCommands(cfg any, slack *string) string {
	var resp string
	err := g.client.Call("Plugin.GetCommands", new(any), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

// Structure for SlackBotPlugin to allow for RPC server and methods
type SlackBotPlugin struct {
	Impl Plugin
}

// Implements Plugin and GRPCPlugin for SlackBot
type SlackBotPluginPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl Plugin
}

func (p *SlackBotPluginPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	s := &CommandRPCServer{Impl: p.Impl}
	server := rpc.NewServer()
	err := server.Register(s)
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (p *SlackBotPluginPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &SlackBotPluginRPC{client: c}, nil
}

func (p *SlackBotPluginPlugin) XXX(b any, c *string) error {
	return nil
}

// RPC Server structure
type CommandRPCServer struct {
	Impl Plugin
}

func (s *CommandRPCServer) GetCommands(args any, resp *string) error {
	if s.Impl == nil {
		return errors.New("implementation is missing for Plugin")
	}

	//	*resp = s.Impl.GetCommands(nil, &client.Slack{})
	return nil
}

func LoadPlugins(b *Bot) Commands {
	commands := Commands{}

	// todo binaries, err := goplugin.Discover("*", path)
	for _, pluginPath := range b.config.Plugins {
		log.Infof("Load plugin %s...", pluginPath)

		c, err := loadPlugin(pluginPath)
		if err != nil {
			fmt.Println(err)
			continue
		}
		s := ""
		pluginCommands := c.GetCommands(b, &s)
		fmt.Println(pluginCommands)
		//commands.Merge(pluginCommands)
	}

	return commands
}

func loadPlugin(pluginPath string) (*SlackBotPluginRPC, error) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin.master",
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
			"command": &SlackBotPluginPlugin{},
		},
		Cmd:    exec.Command(pluginPath),
		Logger: logger,
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense("command")
	if err != nil {
		return nil, err
	}

	return raw.(*SlackBotPluginRPC), nil
}

// Serve the plugin implementation
func ServePlugin(impl Plugin) {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin.sub",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "COMMAND_PLUGIN",
			MagicCookieValue: "bar",
		},
		Plugins: map[string]plugin.Plugin{
			"command": &SlackBotPluginPlugin{Impl: impl},
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})
}
