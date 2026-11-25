package config

import (
	"os"
	"testing"
)

func TestLowResourceParsesFromV8AndLegacyLayouts(t *testing.T) {
	v8, errV8 := ParseConfigBytes([]byte("config-version: 8\nserver:\n  low-resource: true\n  port: 8317\naccess:\n  api-keys: [client]\n"))
	if errV8 != nil {
		t.Fatalf("parse v8: %v", errV8)
	}
	if !v8.LowResource || v8.Port != 8317 {
		t.Fatalf("v8 low-resource = %t port = %d", v8.LowResource, v8.Port)
	}

	legacy, errLegacy := ParseConfigBytes([]byte("low-resource: true\nport: 8318\n"))
	if errLegacy != nil {
		t.Fatalf("parse legacy: %v", errLegacy)
	}
	if !legacy.LowResource || legacy.Port != 8318 {
		t.Fatalf("legacy low-resource = %t port = %d", legacy.LowResource, legacy.Port)
	}
}

func TestRaspberryPiSampleConfigParses(t *testing.T) {
	data, err := os.ReadFile("../../deploy/raspberry-pi/config.yaml")
	if err != nil {
		t.Fatalf("read sample config: %v", err)
	}
	cfg, err := ParseConfigBytes(data)
	if err != nil {
		t.Fatalf("parse sample config: %v", err)
	}
	if !cfg.LowResource || !cfg.CommercialMode || !cfg.RemoteManagement.DisableAutoUpdatePanel {
		t.Fatalf("sample profile = low-resource:%t commercial:%t disable-auto-update:%t", cfg.LowResource, cfg.CommercialMode, cfg.RemoteManagement.DisableAutoUpdatePanel)
	}
	if cfg.RemoteManagement.PanelGitHubRepository != "hansdoebel/Cli-Proxy-API-Dashboard" {
		t.Fatalf("panel repo = %q", cfg.RemoteManagement.PanelGitHubRepository)
	}
	if cfg.AuthAutoRefreshWorkers != 1 || cfg.KVStore.Disabled != true {
		t.Fatalf("workers = %d kv disabled = %t", cfg.AuthAutoRefreshWorkers, cfg.KVStore.Disabled)
	}
}
