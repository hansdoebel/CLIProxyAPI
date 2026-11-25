//go:build notui

package main

import (
	"fmt"
	"os"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/api"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/pluginhost"
)

func startTUI(tuiMode, _, _ bool, _, _, _ string, _ *config.Config, _ *pluginhost.Host, _ []api.ServerOption) (int, bool) {
	if !tuiMode {
		return 0, false
	}
	fmt.Fprintf(os.Stderr, "TUI support was excluded from this build (notui). Start the server without --tui and use the web management panel.\n")
	return 1, true
}
