package pluginmanager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"github.com/Mark7888/go-plugin-tinkering/pkg/shared"
)

const (
	PluginsDir      = "./plugins"
	PluginExtension = ".myext"
)

// LoadedPlugin holds the go-plugin client and factory handle for a running plugin.
type LoadedPlugin struct {
	Name    string
	Client  *plugin.Client
	Factory *shared.FactoryRPCClient
}

// instanceEntry associates a PluginInstance with the plugin that owns it.
type instanceEntry struct {
	pluginName string
	instance   *shared.PluginInstance
}

// Manager manages the lifecycle of all plugins and their instances.
type Manager struct {
	mu        sync.RWMutex
	plugins   map[string]*LoadedPlugin
	instances map[string]*instanceEntry
}

// NewManager creates a new Manager instance.
func NewManager() *Manager {
	return &Manager{
		plugins:   make(map[string]*LoadedPlugin),
		instances: make(map[string]*instanceEntry),
	}
}

// LoadAll reads PluginsDir and loads every file with PluginExtension.
// A warning is logged for any plugin that fails to load.
func (m *Manager) LoadAll() {
	entries, err := os.ReadDir(PluginsDir)
	if err != nil {
		fmt.Printf("[WARN] could not read plugins directory %q: %v\n", PluginsDir, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != PluginExtension {
			continue
		}
		name := entry.Name()[:len(entry.Name())-len(PluginExtension)]
		path := filepath.Join(PluginsDir, entry.Name())
		if err := m.loadPlugin(name, path); err != nil {
			fmt.Printf("[WARN] failed to load plugin %q: %v\n", name, err)
		}
	}
}

// loadPlugin starts a single plugin process and registers its factory.
func (m *Manager) loadPlugin(name, path string) error {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.FactoryPluginMap,
		Cmd:             exec.Command(path),
		Logger: hclog.New(&hclog.LoggerOptions{
			Name:   name,
			Output: io.Discard,
			Level:  hclog.Error,
		}),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("connect: %w", err)
	}

	raw, err := rpcClient.Dispense("plugin")
	if err != nil {
		client.Kill()
		return fmt.Errorf("dispense: %w", err)
	}

	factory, ok := raw.(*shared.FactoryRPCClient)
	if !ok {
		client.Kill()
		return fmt.Errorf("dispensed value is not a *FactoryRPCClient")
	}

	m.mu.Lock()
	m.plugins[name] = &LoadedPlugin{Name: name, Client: client, Factory: factory}
	m.mu.Unlock()

	return nil
}

// UnloadAll destroys all instances, calls DestroyAll on each plugin, and kills processes.
func (m *Manager) UnloadAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear the instance registry — the plugin processes are going away anyway.
	m.instances = make(map[string]*instanceEntry)

	for name, lp := range m.plugins {
		if err := lp.Factory.DestroyAll(); err != nil {
			fmt.Printf("[WARN] plugin %q DestroyAll() error: %v\n", name, err)
		}
		lp.Client.Kill()
		delete(m.plugins, name)
	}
}

// Reload unloads all plugins and then loads them again.
func (m *Manager) Reload() {
	m.UnloadAll()
	m.LoadAll()
}

// List returns the names of all currently loaded plugins.
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	return names
}

// GetFactory returns the FactoryRPCClient for the given plugin name.
func (m *Manager) GetFactory(name string) (*shared.FactoryRPCClient, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lp, ok := m.plugins[name]
	if !ok {
		return nil, false
	}
	return lp.Factory, true
}

// CreateInstance creates a new independent instance of the named plugin and registers it.
func (m *Manager) CreateInstance(pluginName string) (*shared.PluginInstance, error) {
	factory, ok := m.GetFactory(pluginName)
	if !ok {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}

	inst, err := factory.NewInstance()
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	m.instances[inst.ID()] = &instanceEntry{pluginName: pluginName, instance: inst}
	m.mu.Unlock()

	return inst, nil
}

// GetInstance retrieves a registered instance by its UUID.
func (m *Manager) GetInstance(id string) (*shared.PluginInstance, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, ok := m.instances[id]
	if !ok {
		return nil, false
	}
	return entry.instance, true
}

// DestroyInstance destroys an instance by UUID and removes it from the registry.
func (m *Manager) DestroyInstance(id string) error {
	m.mu.Lock()
	entry, ok := m.instances[id]
	if ok {
		delete(m.instances, id)
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("instance not found: %s", id)
	}
	return entry.instance.Destroy()
}

// ListInstances returns the IDs of all live instances for the given plugin.
func (m *Manager) ListInstances(pluginName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var ids []string
	for id, entry := range m.instances {
		if entry.pluginName == pluginName {
			ids = append(ids, id)
		}
	}
	return ids
}
