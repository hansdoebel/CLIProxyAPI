# Raspberry Pi

CLIProxyAPI can run on a Raspberry Pi 2 Model B (1 GB, 32-bit ARMv7) and a Pi Zero 2 W (512 MB, 64-bit ARM). Use the lean build, the sample config, and the systemd unit in `deploy/raspberry-pi/`. A normal `go build` is unchanged.

The numbers below were taken on an x86_64 Linux VM, not on a Pi. Treat them as a comparison of the same binary shapes. A Pi will be slower to start and will have less free RAM for the operating system.

## What uses memory at idle

A default server process, built with symbols stripped, sat at about 59 MB RSS before any of this work. Roughly 40 MB of that was the mapped binary. The rest was heap.

These loops are running even when nobody is calling the API:

- Model catalog refresh every 3 hours. `--local-model` loads the embedded catalogs and does not keep the remote refresh ticker.
- Management panel check against GitHub every 3 hours, unless `disable-auto-update-panel` is set. A missing `management.html` is still downloaded once on first visit.
- Config and auth file watches (`fsnotify`).
- Auth auto-refresh. The default pool is 16 workers.
- Copilot catalog reconcile every minute, with 4 workers.
- Antigravity model scan every minute.
- Usage-queue maintenance.
- mDNS refresh every 15 seconds, only when discovery is enabled.
- Signature-cache cleanup every 10 minutes.
- Optional bbolt KV file (`server.kv-store`). It is on unless disabled.

Request logs, error-log retention, and the in-memory usage queue are already bounded by config (`error-logs-max-files`, `logs-max-total-size-mb`, `redis-usage-queue-retention-seconds`). Usage statistics are off unless enabled. Replay caches were already capped, but the caps were large for a 512 MB board (10240 sessions, up to 16 MB per session, and 256 MB aggregate for some providers).

The heavy linked libraries in a full build include tiktoken encoding tables, Pion WebRTC, Postgres (`pgx`), git storage (`go-git`), S3 storage (`minio`), mDNS (`zeroconf` and `miekg/dns`), the Bubble Tea TUI, uTLS, bbolt, and the Home/Redis client used by cluster mode.

## Lean build

`make build` is a plain `go build`. Pi targets add `-trimpath`, `-ldflags '-s -w'`, `CGO_ENABLED=0`, and these tags:

| Tag | Left out | Trade-off |
| --- | --- | --- |
| `notui` | Bubble Tea management UI | Use the web panel. `--tui` exits with an error. |
| `nowebrtc` | Pion WebRTC media relay | Codex live audio/WebRTC relay is unavailable. SDP forwarding still behaves like the existing unsupported-platform build. |
| `nodiscovery` | mDNS advertise and browse | `discovery.enabled` and `discover` cannot scan the LAN. |
| `noremostores` | Postgres, git, and S3 token stores | `PGSTORE_*`, `GITSTORE_*`, and `OBJECTSTORE_*` fail at startup. File auth storage still works. |
| `notiktoken` | tiktoken tables | Token-count endpoints use about 4 bytes per token instead of model encodings. |

Drop a tag to keep that feature:

```bash
make build-pi LEAN_TAGS='notui,nowebrtc,nodiscovery'
```

```bash
make build-pi-armv7    # linux/arm GOARM=7 -> dist/cli-proxy-api-linux-armv7
make build-pi-arm64    # linux/arm64       -> dist/cli-proxy-api-linux-arm64
make build-pi          # both
```

`CGO_ENABLED=0` produces a static binary and disables Go plugins. Build on the Pi with cgo if you need plugins.

## Low-resource profile

`server.low-resource: true` or `CLIPROXY_LOW_RESOURCE=1` does not change routing. It:

- Caps replay caches at 256 entries, 2 MB per entry, and 8 MB aggregate, and caps each signature-cache group at 128 entries.
- Slows catalog refresh to 12 hours, management-asset checks to 24 hours, discovery refresh to 2 minutes, and Copilot and Antigravity scans to 5 minutes. Loops that have already started keep their original interval, so set this before the process serves traffic.
- Uses 1 auth refresh worker when `auth-auto-refresh-workers` is unset.
- Sets a 192 MB soft heap limit when `GOMEMLIMIT` is unset, and `GOGC=50` when `GOGC` is unset. Values you set in the environment win.

The sample config also turns on settings the flag does not force, because they change behavior:

- `commercial-mode: true` and response compression off, to avoid per-request and encoder buffers.
- `kv-store.disabled: true`, so bbolt is not mapped. Thinking-signature continuity then lasts only until restart.
- Request logs off, error logs capped at 3 files and 8 MB, usage statistics off, usage-queue retention 30 seconds.
- Discovery off, plugins off, `--local-model` in the systemd unit.

`deploy/raspberry-pi/config.yaml` is that profile. `deploy/raspberry-pi/cli-proxy-api.service` runs it with `GOMEMLIMIT=192MiB` and `GOGC=50`. On a Pi 2 you can raise `GOMEMLIMIT` to `384MiB`.

The unit reads `CLIPROXY_CONFIG` (default `/etc/cli-proxy-api/config.yaml`). Copy either `config.yaml` or `config.headless.yaml` to that path, or point the variable at the headless file in a drop-in.

## Management panel offline

The server serves one local file, `management.html`.

1. Download `management.html` from the panel repository you set in `management.panel-github-repository`.
2. Put it in the directory named by `MANAGEMENT_STATIC_PATH`, or in `<config-dir>/static/management.html`.
3. Set `management.disable-auto-update-panel: true`.

With the file present, `GET /management.html` does not contact GitHub. `management.panel-github-repository` accepts `YOUR_GITHUB_USER/YOUR_PANEL_REPO`, `github.com/owner/repo`, or a full GitHub URL. That value is used only when a download actually runs (the file is missing, or auto-update is still on).

## Headless

`deploy/raspberry-pi/config.headless.yaml` is the same low-resource profile with the panel and the management API left off:

- `management.disable-control-panel: true` so `/management.html` is not served.
- `management.secret-key: ""` so `/v0/management` and `/v8/management` are not registered.
- `management.disable-auto-update-panel: true` so the process does not download a panel.

Headless does not shrink the binary. On the lean build, with the same `GOMEMLIMIT` and `GOGC`, idle RSS was 43,896 kB headless and 43,860 kB with the panel profile. That gap is noise. The panel is a separate file, often a few megabytes on disk, and it is not part of the process until a client fetches it. Sending a 3 MiB `management.html` added about 200 kB of process RSS, because the server does not copy the file into the heap. Headless never downloads or serves that file, and it does not register the management routes.

Provider logins are not in the config. On the Pi, run one login flag and then start the server, for example:

```bash
cli-proxy-api --config /etc/cli-proxy-api/config.yaml --codex-device-login --no-browser
```

`--codex-device-login` prints a code and does not open a browser. Other providers have their own `--*-login` flags. You can also copy existing token JSON files into `oauth.auth-dir` (`/var/lib/cli-proxy-api/auths`). Clients still authenticate with `access.api-keys`; replace `YOUR_CLIENT_API_KEY` before exposing the port.

## Measurements

Host: x86_64 Linux VM. Idle RSS is `VmRSS` from `/proc/<pid>/status` about 1.5 seconds after the port accepts. Load RSS is the same after 30 authenticated `GET /v1/models`. Startup is the time until the port accepts. Goroutine totals come from pprof, which itself adds one server. `--local-model` and `disable-auto-update-panel` were set so a catalog or panel download would not dominate the sample. The "before" binary is the tree before these changes.

| Build | File size | Idle RSS | RSS after 30 model lists | Startup | Goroutines |
| --- | ---: | ---: | ---: | ---: | ---: |
| `go build` (before) | 106,314,288 bytes (101.4 MiB) | — | — | — | — |
| stripped full (before) | 80,400,616 bytes (76.7 MiB) | 60,216 kB | 61,472 kB | 0.205 s | 46 |
| stripped full (after, same flags) | 80,409,384 bytes (76.7 MiB) | 59,176 kB | 60,224 kB | 0.152 s | 46 |
| lean amd64, same config as the first runtime row | 43,358,856 bytes (41.4 MiB) | 48,708 kB | 50,180 kB | 0.306 s | 46 |
| lean amd64 + low-resource config, `GOMEMLIMIT=192MiB`, `GOGC=50` | 43,358,856 bytes | 44,204 kB (peak 47,136 kB) | 45,444 kB | 0.201 s | 30 |
| lean `linux/arm` `GOARM=7` | 38,404,258 bytes (36.6 MiB) | not measured on device | — | — | — |
| lean `linux/arm64` | 37,552,290 bytes (35.8 MiB) | not measured on device | — | — | — |

Anonymous RSS on the stripped full build was about 19 MB idle. The lean low-resource process was about 15.4 MB anonymous and 28 MB of mapped file, against 41 MB of mapped file for the full binary.

## What else could be cut

- Home / cluster mode still links the Redis client. A `nohome` tag would remove cluster mode and that client. It is a larger split because Home KV helpers are on the normal cache path.
- uTLS stays in every build so Claude transport fingerprinting still works.
- bbolt stays linked. The sample config turns the file off at runtime.
- Individual providers are not behind tags. Removing one would drop that provider's API.
- `GOMAXPROCS=2` can reduce scheduler threads on a 4-core Pi. It is not set by default because it also caps useful parallelism.
