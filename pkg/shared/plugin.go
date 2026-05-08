package shared

import "github.com/hashicorp/go-plugin"

// Handshake is shared between host and plugins.
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "go-plugin-tinkering",
}

// PluginBase is the interface that every plugin implementation must satisfy.
type PluginBase interface {
	Init() error
	Call(message string) (string, error)
	Shutdown() error
}
