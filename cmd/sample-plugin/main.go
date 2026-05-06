package main

import (
	"github.com/hashicorp/go-plugin"

	"github.com/Mark7888/go-plugin-tinkering/pkg/shared"
)

// SamplePlugin is a trivial implementation of PluginBase.
type SamplePlugin struct{}

// Init performs plugin initialization. This is a no-op for the sample plugin.
func (s *SamplePlugin) Init() error { return nil }

// Call echoes the input message back to the caller.
func (s *SamplePlugin) Call(message string) (string, error) { return message, nil }

// Shutdown performs cleanup when the plugin is unloaded. This is a no-op for the sample plugin.
func (s *SamplePlugin) Shutdown() error { return nil }

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"plugin": &shared.PluginBasePlugin{Impl: &SamplePlugin{}},
		},
	})
}
