// Package lowresource applies memory-conservative runtime limits for small hosts
// such as a Raspberry Pi. The default process behavior is unchanged unless
// server.low-resource is true or CLIPROXY_LOW_RESOURCE is 1 or true.
package lowresource

import (
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/cache"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/managementasset"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/cliproxy"
	log "github.com/sirupsen/logrus"
)

const softHeapLimitBytes = 192 << 20

// Enabled reports whether the low-resource profile should run.
func Enabled(cfg *config.Config) bool {
	if cfg != nil && cfg.LowResource {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("CLIPROXY_LOW_RESOURCE"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// Apply tightens cache caps, slows background refresh loops, and sets a soft
// heap limit when the operator has not already set GOMEMLIMIT or GOGC.
// Explicit auth-auto-refresh-workers values are left alone.
func Apply(cfg *config.Config) {
	if !Enabled(cfg) {
		return
	}
	cache.ApplyLowResourceLimits()
	registry.SetModelsRefreshInterval(12 * time.Hour)
	managementasset.SetUpdateCheckInterval(24 * time.Hour)
	cliproxy.SetDiscoveryRefreshInterval(2 * time.Minute)
	cliproxy.SetCopilotCatalogRefreshInterval(5 * time.Minute)
	cliproxy.SetAntigravityRefreshInterval(5 * time.Minute)
	if cfg != nil && cfg.AuthAutoRefreshWorkers <= 0 {
		cfg.AuthAutoRefreshWorkers = 1
	}
	if strings.TrimSpace(os.Getenv("GOMEMLIMIT")) == "" {
		debug.SetMemoryLimit(softHeapLimitBytes)
		log.Infof("low-resource: GOMEMLIMIT is unset, soft heap limit is %d MiB", softHeapLimitBytes>>20)
	}
	if strings.TrimSpace(os.Getenv("GOGC")) == "" {
		debug.SetGCPercent(50)
		log.Info("low-resource: GOGC is unset, GC percent is 50")
	}
	log.Info("low-resource profile enabled: replay caches capped, catalog refresh 12h, management asset check 24h, discovery refresh 2m")
}
