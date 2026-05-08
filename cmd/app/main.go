// @title           Go Plugin Tinkering API
// @version         1.0
// @description     REST API for managing and calling hashicorp/go-plugin plugins
// @host            localhost:8080
// @BasePath        /api/v1
package main

import (
	"log"

	_ "github.com/Mark7888/go-plugin-tinkering/docs"
	"github.com/Mark7888/go-plugin-tinkering/internal/api"
	"github.com/Mark7888/go-plugin-tinkering/pkg/pluginmanager"
)

func main() {
	m := pluginmanager.NewManager()
	m.LoadAll()

	r := api.NewRouter(m)

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
