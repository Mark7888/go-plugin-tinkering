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

// LoadedPlugin holds the client and dispensed interface for a running plugin.
type LoadedPlugin struct {
	Name   string
	Client *plugin.Client
	Plugin shared.PluginBase
}

// Manager manages the lifecycle of all plugins.
type Manager struct {
	mu      sync.RWMutex
	plugins map[string]*LoadedPlugin
}

// NewManager creates a new Manager instance.
func NewManager() *Manager {
	return &Manager{
		plugins: make(map[string]*LoadedPlugin),
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

// loadPlugin starts a single plugin process and calls Init().
func (m *Manager) loadPlugin(name, path string) error {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.PluginMap,
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

	p, ok := raw.(shared.PluginBase)
	if !ok {
		client.Kill()
		return fmt.Errorf("dispensed value does not implement PluginBase")
	}

	if err := p.Init(); err != nil {
		client.Kill()
		return fmt.Errorf("Init(): %w", err)
	}

	m.mu.Lock()
	m.plugins[name] = &LoadedPlugin{Name: name, Client: client, Plugin: p}
	m.mu.Unlock()

	return nil
}

// UnloadAll calls Shutdown() on every plugin and kills its process.
func (m *Manager) UnloadAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, lp := range m.plugins {
		if err := lp.Plugin.Shutdown(); err != nil {
			fmt.Printf("[WARN] plugin %q Shutdown() error: %v\n", name, err)
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

// GetPlugin returns the PluginBase implementation for the given name.
func (m *Manager) GetPlugin(name string) (shared.PluginBase, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	lp, ok := m.plugins[name]
	if !ok {
		return nil, false
	}
	return lp.Plugin, true
}
