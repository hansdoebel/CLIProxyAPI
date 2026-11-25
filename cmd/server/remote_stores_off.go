//go:build noremostores

package main

import (
	"fmt"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	log "github.com/sirupsen/logrus"
)

type remoteStoreResult struct {
	cfg            *config.Config
	configFilePath string
	register       func()
}

func openRemoteStore(env remoteStoreEnv, _ bool, _, _ string) (remoteStoreResult, error) {
	if env.usePostgres || env.useObject || env.useGit {
		err := fmt.Errorf("postgres, git, and object-store backends were excluded from this build (noremostores)")
		log.Errorf("%v", err)
		return remoteStoreResult{}, err
	}
	return remoteStoreResult{}, nil
}
