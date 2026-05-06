package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/Mark7888/go-plugin-tinkering/pkg/pluginmanager"
)

// NewRouter creates and returns a configured Gin engine.
func NewRouter(m *pluginmanager.Manager) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/plugins/load", LoadPlugins(m))
		v1.POST("/plugins/unload", UnloadPlugins(m))
		v1.POST("/plugins/reload", ReloadPlugins(m))
		v1.GET("/plugins", ListPlugins(m))
		v1.POST("/plugins/:name/call", CallPlugin(m))
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
