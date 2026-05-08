package main

import (
	"fmt"

	"github.com/hashicorp/go-plugin"

	"github.com/Mark7888/go-plugin-tinkering/pkg/shared"
)

// CounterPlugin maintains an independent call counter per instance.
// Each instance created from this plugin binary has its own counter that
// starts at zero — demonstrating that instances are fully isolated.
type CounterPlugin struct {
	count int
}

// Init resets the counter when the instance is created.
func (c *CounterPlugin) Init() error {
	c.count = 0
	return nil
}

// Call increments the counter and returns the message together with the current count.
func (c *CounterPlugin) Call(message string) (string, error) {
	c.count++
	return fmt.Sprintf("[counter] %s (call #%d)", message, c.count), nil
}

// Shutdown is called when the instance is destroyed.
func (c *CounterPlugin) Shutdown() error { return nil }

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins:         shared.NewFactoryPluginMap(func() shared.PluginBase { return &CounterPlugin{} }),
	})
}
