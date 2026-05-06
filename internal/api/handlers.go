package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Mark7888/go-plugin-tinkering/pkg/pluginmanager"
)

// LoadPlugins godoc
// @Summary      Load all plugins
// @Description  Loads all plugins from the plugins directory
// @Tags         plugins
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /plugins/load [post]
func LoadPlugins(m *pluginmanager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.LoadAll()
		c.JSON(http.StatusOK, gin.H{"status": "loaded"})
	}
}

// UnloadPlugins godoc
// @Summary      Unload all plugins
// @Description  Shuts down and unloads all currently loaded plugins
// @Tags         plugins
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /plugins/unload [post]
func UnloadPlugins(m *pluginmanager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.UnloadAll()
		c.JSON(http.StatusOK, gin.H{"status": "unloaded"})
	}
}

// ReloadPlugins godoc
// @Summary      Reload all plugins
// @Description  Unloads and then reloads all plugins from the plugins directory
// @Tags         plugins
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /plugins/reload [post]
func ReloadPlugins(m *pluginmanager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		m.Reload()
		c.JSON(http.StatusOK, gin.H{"status": "reloaded"})
	}
}

// ListPlugins godoc
// @Summary      List loaded plugins
// @Description  Returns a list of all currently loaded plugin names
// @Tags         plugins
// @Produce      json
// @Success      200  {object}  map[string][]string
// @Router       /plugins [get]
func ListPlugins(m *pluginmanager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"plugins": m.List()})
	}
}

// callRequest is the request body for the CallPlugin endpoint.
type callRequest struct {
	Message string `json:"message" binding:"required"`
}

// CallPlugin godoc
// @Summary      Call a plugin
// @Description  Calls the named plugin with the provided message and returns the result
// @Tags         plugins
// @Accept       json
// @Produce      json
// @Param        name  path      string      true  "Plugin name"
// @Param        body  body      callRequest true  "Message payload"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /plugins/{name}/call [post]
func CallPlugin(m *pluginmanager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		p, ok := m.GetPlugin(name)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found: " + name})
			return
		}

		var req callRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := p.Call(req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": result})
	}
}
