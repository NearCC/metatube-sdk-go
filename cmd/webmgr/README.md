# webmgr

A standalone web UI for `metatube-sdk-go` — a movie/actor metadata
management frontend (Vue 3) and the SDK's REST API (for Jellyfin / Emby
plugins) bundled into a **single Go binary**. There is no separate
`metatube-server` subprocess; the SDK engine is linked in directly and
shared between the two HTTP listeners.

```
                ┌──────────────────────────────┐
   browser ───► │  :8081  webmgr UI (Gin)      │ ──►  sdkEngine (in-process)
                │                              │
  Jellyfin ───► │  :8080  metatube REST API    │ ──►  sdkEngine (in-process)
                └──────────────────────────────┘
                          single process, one SDK engine
```

- **:8081** — web UI: login, users, favorites, history, follows, image cache.
- **:8080** — the SDK's own REST API (`route.New(...)`), unchanged from
  the upstream `metatube-server` so Jellyfin/Emby plugins keep working
  with no URL change. Auth via `X-API-Token` (set with `METATUBE_TOKEN`).

## Build

```sh
# frontend (Vite) is built by the first `go build` thanks to //go:embed
# all:dist, but the dist/ tree must already exist. To (re)build the UI:
cd cmd/webmgr/frontend && npm ci && npm run build && cd -

# then build the binary
go build -o webmgr ./cmd/webmgr
```

Cross-compiles cleanly to Windows, macOS, Linux on amd64/arm64.

## Run

```powershell
# Windows
cd cmd\webmgr
..\..\webmgr.exe
```

```sh
# macOS / Linux
./webmgr
```

The binary reads `config.yaml` from its working directory. By default:

- UI:    <http://localhost:8081>
- API:   <http://localhost:8080>

Default credentials: **`admin` / `admin`** (change after first login).

## Configuration

`cmd/webmgr/config.yaml`:

```yaml
app:
  host: "0.0.0.0"
  port: 8081
  data_dir: "./data"

metatube:
  mode: "embedded"        # currently the only supported mode
  token: ""               # X-API-Token for the REST API; empty = open
  movie_provider: "JavBus"
  api_host: "0.0.0.0"
  api_port: 8080

log:
  filename: "webmgr.log"  # written to <data_dir>/logs/
  max_size_mb: 16         # rotate when the file exceeds this; 0 = no rotation
  max_files: 5            # keep webmgr.log plus webmgr.log.1 .. .N
```

Every field can be overridden by an environment variable (env wins):

| Env var                      | Default   | Notes                                |
|------------------------------|-----------|--------------------------------------|
| `APP_HOST`                   | `0.0.0.0` | UI bind address                      |
| `APP_PORT`                   | `8081`    | UI port                              |
| `APP_DATA_DIR`               | `./data`  | SQLite + image cache + log file      |
| `METATUBE_TOKEN`             | (empty)   | Require `X-API-Token` on :8080       |
| `METATUBE_MOVIE_PROVIDER`    | `JavBus`  | Default provider                     |
| `METATUBE_API_HOST`          | `0.0.0.0` | REST API bind address                |
| `METATUBE_API_PORT`          | `8080`    | REST API port (Jellyfin/Emby)        |
| `LOG_FILENAME`               | `webmgr.log` | Log file name (inside data_dir/logs/) |
| `LOG_MAX_SIZE_MB`            | `16`      | Rotate above this size; 0 disables   |
| `LOG_MAX_FILES`              | `5`       | Number of rotated files to keep      |
| `MT_MOVIE_PROVIDER_PROXY`    | (empty)   | Per-provider proxy (SDK `ProxySetter`) |
| `MT_PROVIDER_PROXY`          | (empty)   | Global per-provider proxy            |
| `HTTP_PROXY`                 | (empty)   | Go stdlib `ProxyFromEnvironment`     |
| `HTTPS_PROXY`                | (empty)   | Go stdlib `ProxyFromEnvironment`     |

## Logging

Logs go to **both** stdout (so `docker logs` works) and a rotating file at
`<data_dir>/logs/<log.filename>`:

```
data/
├── data.db
└── logs/
    ├── webmgr.log      ← current
    ├── webmgr.log.1    ← previous rotations
    └── webmgr.log.2
```

Set `log.max_size_mb: 0` to disable rotation and keep a single growing
file.

## Docker

### Prerequisites

- Docker 20.10+ (28+ recommended)
- ~4 GB free disk for the build cache (node_modules + Go module cache + image layers)
- ~2 GB RAM during the `golang` build stage

### End-to-end on a fresh machine

```sh
git clone https://github.com/NearCC/metatube-sdk-go.git
cd metatube-sdk-go
git switch feat/embed-web-manager
git pull

docker build -f Dockerfile.webmgr -t metatube-webmgr:local .
```

Or with the bundled compose file (builds and starts in one step):

```sh
docker compose -f docker-compose.webmgr.yaml up -d
```

The image exposes both ports. The data volume at `/app/data` persists
the SQLite database and image cache.

After the container is up:

- UI:  <http://localhost:8081>  (default `admin` / `admin`)
- API: <http://localhost:8080>

### Build time

- First build: 5–10 min (downloads `node_modules` and Go module cache).
- Subsequent builds: 10–30 s (Docker layer cache hits on `npm ci` and
  `go build`).

To rebuild from scratch:

```sh
docker build --no-cache -f Dockerfile.webmgr -t metatube-webmgr:local .
```

## Network / proxy

`webmgr` uses the SDK's own HTTP client for all provider requests, so it
honors Go's `http.ProxyFromEnvironment`:

- `HTTP_PROXY` / `HTTPS_PROXY` for any provider that uses the default
  `http.Transport`.
- `MT_MOVIE_PROVIDER_PROXY` / `MT_PROVIDER_PROXY` for providers that
  implement the SDK's `ProxySetter` interface (more precise, per-provider).

If you can reach providers from `metatube-server` on the same machine,
this binary will too — no extra config.

## Project layout

```
cmd/webmgr/
├── main.go                 # entry point, two http.Server, embeds dist/
├── config.yaml
├── go.mod                  # nested module, replace ../.. -> SDK root
├── Dockerfile.webmgr       # node + golang + debian-slim, multi-stage
├── frontend/               # Vue 3 + Vite source
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
└── dist/                   # vite build output (embedded into binary)

internal/webmgr/            # SDK-internal packages used only by webmgr
├── config/                 # config loading + env overrides
├── handler/                # Gin HTTP handlers
├── service/                # business logic
├── repository/             # GORM repos
├── model/                  # GORM models
├── middleware/             # auth middleware
├── metatube/               # SDK engine adapter (replaces HTTP client)
└── util/                   # password hashing etc.
```

## Differences from the standalone `metatube-web-manager`

This binary is a re-integration of
[metatube-web-manager](https://github.com/metatube-community/metatube-web-manager)
into the SDK repo. Compared to the standalone project:

- **No `metatube-server` subprocess.** The SDK engine is linked in.
  The `embedded` / `external` mode switch and the cross-platform process
  manager are gone — only the embedded mode exists.
- **Port 8080 is the REST API, not the UI.** This matches the original
  `metatube-server` default, so existing Jellyfin/Emby setups keep
  working with the same URL.
- **Same `config.yaml` shape and env vars** as the standalone project,
  with two additions: `metatube.api_host` / `metatube.api_port` (and
  `METATUBE_API_HOST` / `METATUBE_API_PORT`).
- **Single `go.mod` strategy.** The webmgr binary lives in a nested
  module (`cmd/webmgr/go.mod`) with a `replace` directive pointing at
  the SDK root, mirroring the layout of `cmd/metatube-sort/`.
