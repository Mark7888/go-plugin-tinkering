package shared

import (
	"fmt"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// Handshake is shared between host and plugins.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "go-plugin-tinkering",
}

// PluginMap is used by the host to know which plugins to dispense.
var PluginMap = map[string]plugin.Plugin{
	"plugin": &PluginBasePlugin{},
}

// PluginBase is the interface that every plugin must implement.
type PluginBase interface {
	Init() error
	Call(message string) (string, error)
	Shutdown() error
}

// CallResponse carries the result and error string from an RPC call.
type CallResponse struct {
	Result string
	Error  string
}

// PluginBasePlugin implements plugin.Plugin and bridges the RPC layer.
type PluginBasePlugin struct {
	Impl PluginBase
}

func (p *PluginBasePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginBaseRPCServer{Impl: p.Impl}, nil
}

func (p *PluginBasePlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginBaseRPCClient{client: c}, nil
}

// PluginBaseRPCClient is used by the host to call the plugin over RPC.
type PluginBaseRPCClient struct {
	client *rpc.Client
}

func (c *PluginBaseRPCClient) Init() error {
	var resp string
	err := c.client.Call("Plugin.Init", new(interface{}), &resp)
	if err != nil {
		return err
	}
	if resp != "" {
		return fmt.Errorf("%s", resp)
	}
	return nil
}

func (c *PluginBaseRPCClient) Call(message string) (string, error) {
	var resp CallResponse
	err := c.client.Call("Plugin.Call", message, &resp)
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf("%s", resp.Error)
	}
	return resp.Result, nil
}

func (c *PluginBaseRPCClient) Shutdown() error {
	var resp string
	err := c.client.Call("Plugin.Shutdown", new(interface{}), &resp)
	if err != nil {
		return err
	}
	if resp != "" {
		return fmt.Errorf("%s", resp)
	}
	return nil
}

// PluginBaseRPCServer is used by the plugin binary to serve RPC requests.
type PluginBaseRPCServer struct {
	Impl PluginBase
}

func (s *PluginBaseRPCServer) Init(args interface{}, resp *string) error {
	err := s.Impl.Init()
	if err != nil {
		*resp = err.Error()
	}
	return nil
}

func (s *PluginBaseRPCServer) Call(args string, resp *CallResponse) error {
	result, err := s.Impl.Call(args)
	resp.Result = result
	if err != nil {
		resp.Error = err.Error()
	}
	return nil
}

func (s *PluginBaseRPCServer) Shutdown(args interface{}, resp *string) error {
	err := s.Impl.Shutdown()
	if err != nil {
		*resp = err.Error()
	}
	return nil
}
