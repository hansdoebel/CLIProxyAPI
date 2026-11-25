# Default build matches `go build` and keeps every integration.
# Pi targets add -trimpath, strip symbols, and optional build tags.
# Override LEAN_TAGS to keep a feature, for example:
#   make build-pi LEAN_TAGS='notui,nowebrtc,nodiscovery'

BINARY ?= cli-proxy-api
DIST ?= dist
LEAN_TAGS ?= notui,nowebrtc,nodiscovery,noremostores,notiktoken
RELEASE_LDFLAGS := -s -w

.PHONY: build build-stripped build-pi build-pi-armv7 build-pi-arm64

build:
	go build -o $(BINARY) ./cmd/server

build-stripped:
	go build -trimpath -ldflags '$(RELEASE_LDFLAGS)' -o $(BINARY) ./cmd/server

build-pi-armv7:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -tags '$(LEAN_TAGS)' -ldflags '$(RELEASE_LDFLAGS)' -o $(DIST)/cli-proxy-api-linux-armv7 ./cmd/server

build-pi-arm64:
	mkdir -p $(DIST)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -tags '$(LEAN_TAGS)' -ldflags '$(RELEASE_LDFLAGS)' -o $(DIST)/cli-proxy-api-linux-arm64 ./cmd/server

build-pi: build-pi-armv7 build-pi-arm64
