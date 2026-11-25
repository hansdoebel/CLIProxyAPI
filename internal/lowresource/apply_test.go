package lowresource

import (
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/cache"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
)

func TestEnabledRequiresFlagOrEnvironment(t *testing.T) {
	if Enabled(nil) {
		t.Fatal("nil config enabled the profile")
	}
	if Enabled(&config.Config{}) {
		t.Fatal("default config enabled the profile")
	}
	if !Enabled(&config.Config{LowResource: true}) {
		t.Fatal("low-resource config did not enable the profile")
	}
	t.Setenv("CLIPROXY_LOW_RESOURCE", "1")
	if !Enabled(nil) {
		t.Fatal("CLIPROXY_LOW_RESOURCE=1 did not enable the profile")
	}
}

func TestApplyCapsReplayCachesAndRefreshWorkers(t *testing.T) {
	t.Setenv("GOMEMLIMIT", "256MiB")
	t.Setenv("GOGC", "100")
	cfg := &config.Config{LowResource: true}
	Apply(cfg)
	if cfg.AuthAutoRefreshWorkers != 1 {
		t.Fatalf("auth workers = %d, want 1", cfg.AuthAutoRefreshWorkers)
	}
	if cache.CodexReasoningReplayCacheMaxEntries != 10240 {
		t.Fatalf("default entry cap changed to %d", cache.CodexReasoningReplayCacheMaxEntries)
	}
	explicit := &config.Config{LowResource: true, AuthAutoRefreshWorkers: 4}
	Apply(explicit)
	if explicit.AuthAutoRefreshWorkers != 4 {
		t.Fatalf("explicit auth workers = %d, want 4", explicit.AuthAutoRefreshWorkers)
	}
}
