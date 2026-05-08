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
// @Description  Destroys all instances, shuts down, and unloads all currently loaded plugins
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

// CreateInstance godoc
// @Summary      Create a plugin instance
// @Description  Creates a new independent instance of the named plugin with its own internal state
// @Tags         instances
// @Produce      json
// @Param        name  path      string  true  "Plugin name"
// @Success      201   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /plugins/{name}/instances [post]
func CreateInstance(m *pluginmanager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		if _, ok := m.GetFactory(name); !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found: " + name})
			return
		}

		inst, err := m.CreateInstance(name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"instance_id": inst.ID()})
	}
}

// ListInstances godoc
// @Summary      List plugin instances
// @Description  Returns the IDs of all live instances for the named plugin
// @Tags         instances
// @Produce      json
// @Param        name  path      string  true  "Plugin name"
// @Success      200   {object}  map[string][]string
// @Failure      404   {object}  map[string]string
// @Router       /plugins/{name}/instances [get]
func ListInstances(m *pluginmanager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		if _, ok := m.GetFactory(name); !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "plugin not found: " + name})
			return
		}

		c.JSON(http.StatusOK, gin.H{"instances": m.ListInstances(name)})
	}
}

// callRequest is the request body for the CallInstance endpoint.
type callRequest struct {
	Message string `json:"message" binding:"required"`
}

// CallInstance godoc
// @Summary      Call a plugin instance
// @Description  Calls the identified instance with the provided message and returns the result
// @Tags         instances
// @Accept       json
// @Produce      json
// @Param        name  path      string      true  "Plugin name"
// @Param        id    path      string      true  "Instance ID"
// @Param        body  body      callRequest true  "Message payload"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /plugins/{name}/instances/{id}/call [post]
func CallInstance(m *pluginmanager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		inst, ok := m.GetInstance(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "instance not found: " + id})
			return
		}

		var req callRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := inst.Call(req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": result})
	}
}

// DestroyInstance godoc
// @Summary      Destroy a plugin instance
// @Description  Calls Shutdown on the identified instance and removes it
// @Tags         instances
// @Produce      json
// @Param        name  path      string  true  "Plugin name"
// @Param        id    path      string  true  "Instance ID"
// @Success      200   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /plugins/{name}/instances/{id} [delete]
func DestroyInstance(m *pluginmanager.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := m.DestroyInstance(id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "destroyed"})
	}
}
