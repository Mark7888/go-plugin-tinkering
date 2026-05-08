package shared

import (
	"crypto/rand"
	"fmt"
	"net/rpc"
	"strings"
	"sync"

	"github.com/hashicorp/go-plugin"
)

// FactoryPluginMap is the plugin map used by the host application when loading plugins.
var FactoryPluginMap = map[string]plugin.Plugin{
	"plugin": &FactoryPlugin{},
}

// NewFactoryPluginMap returns a plugin map for use inside a plugin binary.
// The constructor is called each time a new independent instance is requested.
func NewFactoryPluginMap(constructor func() PluginBase) map[string]plugin.Plugin {
	return map[string]plugin.Plugin{
		"plugin": &FactoryPlugin{constructor: constructor},
	}
}

// ---- internal RPC arg/response types ----

type createResp struct {
	UUID  string
	Error string
}

type factoryCallArgs struct {
	UUID    string
	Message string
}

type factoryCallResp struct {
	Result string
	Error  string
}

type destroyArgs struct {
	UUID string
}

// ---- FactoryPlugin: implements plugin.Plugin ----

// FactoryPlugin bridges the RPC layer for the factory pattern.
// On the plugin binary side it holds a constructor; on the host side it is empty.
type FactoryPlugin struct {
	constructor func() PluginBase
}

func (p *FactoryPlugin) Server(_ *plugin.MuxBroker) (interface{}, error) {
	return &factoryRPCServer{
		constructor: p.constructor,
		instances:   make(map[string]PluginBase),
	}, nil
}

func (p *FactoryPlugin) Client(_ *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &FactoryRPCClient{client: c}, nil
}

// ---- factoryRPCServer: runs inside the plugin binary ----

// factoryRPCServer manages PluginBase instances on behalf of the factory.
// It is registered with net/rpc under the name "Plugin" by go-plugin.
type factoryRPCServer struct {
	constructor func() PluginBase
	mu          sync.RWMutex
	instances   map[string]PluginBase
}

// Create creates a new PluginBase instance, calls Init, and returns its UUID.
func (s *factoryRPCServer) Create(_ *struct{}, resp *createResp) error {
	inst := s.constructor()
	if err := inst.Init(); err != nil {
		resp.Error = err.Error()
		return nil
	}
	id, err := newUUID()
	if err != nil {
		resp.Error = err.Error()
		return nil
	}
	s.mu.Lock()
	s.instances[id] = inst
	s.mu.Unlock()
	resp.UUID = id
	return nil
}

// Call routes a message to the instance identified by UUID.
func (s *factoryRPCServer) Call(args *factoryCallArgs, resp *factoryCallResp) error {
	s.mu.RLock()
	inst, ok := s.instances[args.UUID]
	s.mu.RUnlock()
	if !ok {
		resp.Error = "instance not found: " + args.UUID
		return nil
	}
	result, err := inst.Call(args.Message)
	resp.Result = result
	if err != nil {
		resp.Error = err.Error()
	}
	return nil
}

// Destroy calls Shutdown on the identified instance and removes it from the registry.
func (s *factoryRPCServer) Destroy(args *destroyArgs, resp *string) error {
	s.mu.Lock()
	inst, ok := s.instances[args.UUID]
	if ok {
		delete(s.instances, args.UUID)
	}
	s.mu.Unlock()
	if ok {
		if err := inst.Shutdown(); err != nil {
			*resp = err.Error()
		}
	}
	return nil
}

// DestroyAll shuts down every instance in the plugin process.
// All Shutdown errors are collected; the combined message is returned.
func (s *factoryRPCServer) DestroyAll(_ *struct{}, resp *string) error {
	s.mu.Lock()
	instances := s.instances
	s.instances = make(map[string]PluginBase)
	s.mu.Unlock()
	var errs []string
	for _, inst := range instances {
		if err := inst.Shutdown(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		*resp = strings.Join(errs, "; ")
	}
	return nil
}

// ---- FactoryRPCClient: runs in the host application ----

// FactoryRPCClient is the host-side handle to a loaded plugin process.
// Use NewInstance to create independent, stateful plugin instances.
type FactoryRPCClient struct {
	client *rpc.Client
}

// NewInstance creates a new plugin instance and returns a handle to it.
// Each call returns an independent instance with its own internal state.
func (f *FactoryRPCClient) NewInstance() (*PluginInstance, error) {
	var resp createResp
	if err := f.client.Call("Plugin.Create", &struct{}{}, &resp); err != nil {
		return nil, err
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	return &PluginInstance{id: resp.UUID, factory: f}, nil
}

// DestroyAll shuts down every instance managed by this plugin process.
func (f *FactoryRPCClient) DestroyAll() error {
	var resp string
	if err := f.client.Call("Plugin.DestroyAll", &struct{}{}, &resp); err != nil {
		return err
	}
	if resp != "" {
		return fmt.Errorf("%s", resp)
	}
	return nil
}

// ---- PluginInstance: a handle to one specific instance ----

// PluginInstance is a handle to a single stateful instance inside a plugin process.
// Instances created from the same plugin binary are fully independent of each other.
type PluginInstance struct {
	id      string
	factory *FactoryRPCClient
}

// ID returns the unique identifier of this instance.
func (i *PluginInstance) ID() string { return i.id }

// Call forwards a message to this instance and returns its response.
func (i *PluginInstance) Call(message string) (string, error) {
	var resp factoryCallResp
	err := i.factory.client.Call("Plugin.Call", &factoryCallArgs{UUID: i.id, Message: message}, &resp)
	if err != nil {
		return "", err
	}
	if resp.Error != "" {
		return "", fmt.Errorf("%s", resp.Error)
	}
	return resp.Result, nil
}

// Destroy calls Shutdown on this instance and removes it from the plugin process.
func (i *PluginInstance) Destroy() error {
	var resp string
	err := i.factory.client.Call("Plugin.Destroy", &destroyArgs{UUID: i.id}, &resp)
	if err != nil {
		return err
	}
	if resp != "" {
		return fmt.Errorf("%s", resp)
	}
	return nil
}

// newUUID generates a random UUID v4 and returns an error if the random source fails.
// Format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx  (8-4-4-4-12 hex digits)
// where the version nibble is 0x4 and the variant bits are 0b10xx.
func newUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
