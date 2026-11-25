//go:build !notui

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/api"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/cmd"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/logging"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/managementasset"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/pluginhost"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/registry"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/tui"
	log "github.com/sirupsen/logrus"
)

func startTUI(tuiMode, standalone, localModel bool, configFilePath, managementBaseURL, password string, cfg *config.Config, pluginHost *pluginhost.Host, serverOptions []api.ServerOption) (int, bool) {
	if !tuiMode {
		return 0, false
	}
	if standalone {
		// Standalone mode: start an embedded local server and connect TUI client to it.
		managementasset.StartAutoUpdater(context.Background(), configFilePath)
		registry.SetLocalModelCatalogs(localModel)
		hook := tui.NewLogHook(2000)
		hook.SetFormatter(&logging.LogFormatter{})
		log.AddHook(hook)

		origStdout := os.Stdout
		origStderr := os.Stderr
		origLogOutput := log.StandardLogger().Out
		log.SetOutput(io.Discard)

		devNull, errOpenDevNull := os.Open(os.DevNull)
		if errOpenDevNull == nil {
			os.Stdout = devNull
			os.Stderr = devNull
		}

		restoreIO := func() {
			os.Stdout = origStdout
			os.Stderr = origStderr
			log.SetOutput(origLogOutput)
			if devNull != nil {
				_ = devNull.Close()
			}
		}

		localMgmtPassword := fmt.Sprintf("tui-%d-%d", os.Getpid(), time.Now().UnixNano())
		if password == "" {
			password = localMgmtPassword
		}

		cancel, done := cmd.StartServiceBackgroundWithPluginHost(cfg, configFilePath, password, pluginHost, serverOptions...)

		client := tui.NewClient(cfg.Port, password)
		ready := false
		backoff := 100 * time.Millisecond
		for i := 0; i < 30; i++ {
			select {
			case <-done:
				i = 30
			default:
			}
			if i >= 30 {
				break
			}
			if _, errGetConfig := client.GetConfig(); errGetConfig == nil {
				ready = true
				break
			}
			select {
			case <-done:
				i = 30
			case <-time.After(backoff):
			}
			if backoff < time.Second {
				backoff = time.Duration(float64(backoff) * 1.5)
			}
		}

		if !ready {
			restoreIO()
			cancel()
			<-done
			fmt.Fprintf(os.Stderr, "TUI error: embedded server is not ready\n")
			return 1, true
		}

		var errTUI error
		if errRun := tui.Run(cfg.Port, password, hook, origStdout); errRun != nil {
			restoreIO()
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", errRun)
			errTUI = errRun
		} else {
			restoreIO()
		}

		cancel()
		<-done
		if errTUI != nil {
			return 1, true
		}
	} else {
		// Default TUI mode: pure management client.
		// The proxy server must already be running (locally or remotely).
		baseURL := resolveManagementBaseURL(managementBaseURL, cfg)
		if errRun := tui.RunWithBaseURL(baseURL, password, nil, os.Stdout); errRun != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", errRun)
			return 1, true
		}
	}
	return 0, true
}
