# Go Plugin Tinkering

A Go application that demonstrates loading and managing external plugins via [hashicorp/go-plugin](https://github.com/hashicorp/go-plugin), exposed through a REST API built with [gin-gonic/gin](https://github.com/gin-gonic/gin).

---

## What it does

- Discovers and loads every `*.myext` binary from the `plugins/` directory on startup.
- Manages the plugin lifecycle: load, unload, reload.
- Exposes a REST API to interact with loaded plugins at runtime.
- Serves Swagger UI documentation at `/swagger/index.html`.

---

## Prerequisites

- Go 1.21+
- `swag` CLI (`go install github.com/swaggo/swag/cmd/swag@latest`)

---

## Build & Run

```bash
# Build both the app and the sample plugin, then start the server
make run

# Or step by step:
make swagger      # regenerate docs/
make build-all    # compile bin/app and plugins/sample.myext
./bin/app
```

The server listens on **http://localhost:8080**.

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/plugins/load` | Load all plugins from `plugins/` |
| `POST` | `/api/v1/plugins/unload` | Unload all plugins |
| `POST` | `/api/v1/plugins/reload` | Reload all plugins |
| `GET`  | `/api/v1/plugins` | List loaded plugin names |
| `POST` | `/api/v1/plugins/:name/call` | Call a plugin with `{"message": "..."}` |

### Example

```bash
curl -s http://localhost:8080/api/v1/plugins
# {"plugins":["sample"]}

curl -s -X POST http://localhost:8080/api/v1/plugins/sample/call \
  -H 'Content-Type: application/json' \
  -d '{"message":"hello"}'
# {"result":"hello"}
```

---

## Swagger UI

Open **http://localhost:8080/swagger/index.html** in your browser.

---

## Adding a New Plugin

1. Create a new Go package under `cmd/my-plugin/main.go`.
2. Implement the `shared.PluginBase` interface (`Init`, `Call`, `Shutdown`).
3. Call `plugin.Serve` with `shared.Handshake` and a `shared.PluginBasePlugin{Impl: &MyPlugin{}}`.
4. Build it to `plugins/my-plugin.myext`:
   ```bash
   go build -o plugins/my-plugin.myext ./cmd/my-plugin/
   ```
5. Restart the app (or call `POST /api/v1/plugins/reload`).
Testing hashicorp/go-plugin in an example project
