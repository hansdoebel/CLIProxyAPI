//go:build !noremostores

package main

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v8/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v8/internal/store"
	sdkAuth "github.com/router-for-me/CLIProxyAPI/v8/sdk/auth"
	log "github.com/sirupsen/logrus"
)

type remoteStoreResult struct {
	cfg            *config.Config
	configFilePath string
	register       func()
}

func openRemoteStore(env remoteStoreEnv, isCloudDeploy bool, wd, writableBase string) (remoteStoreResult, error) {
	var result remoteStoreResult
	if env.usePostgres {
		localPath := env.pgLocal
		if localPath == "" {
			localPath = wd
		}
		localPath = filepath.Join(localPath, "pgstore")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		pgStoreInst, err := store.NewPostgresStore(ctx, store.PostgresStoreConfig{
			DSN:            env.pgDSN,
			Schema:         env.pgSchema,
			SpoolDir:       localPath,
			OptionalConfig: isCloudDeploy,
		})
		cancel()
		if err != nil {
			log.Errorf("failed to initialize postgres token store: %v", err)
			return remoteStoreResult{}, err
		}
		examplePath := filepath.Join(wd, "config.example.yaml")
		cfg, err := pgStoreInst.BootstrapConfig(context.Background(), examplePath)
		if err != nil {
			log.Errorf("failed to bootstrap postgres-backed config: %v", err)
			return remoteStoreResult{}, err
		}
		cfg.AuthDir = pgStoreInst.AuthDir()
		log.Infof("postgres-backed token store enabled, workspace path: %s", pgStoreInst.WorkDir())
		result.cfg = cfg
		result.configFilePath = pgStoreInst.ConfigPath()
		result.register = func() { sdkAuth.RegisterTokenStore(pgStoreInst) }
		return result, nil
	}
	if env.useObject {
		localPath := env.objectLocal
		if localPath == "" {
			if writableBase != "" {
				localPath = writableBase
			} else {
				localPath = wd
			}
		}
		objectStoreRoot := filepath.Join(localPath, "objectstore")
		resolvedEndpoint := strings.TrimSpace(env.objectEndpoint)
		useSSL := true
		if strings.Contains(resolvedEndpoint, "://") {
			parsed, errParse := url.Parse(resolvedEndpoint)
			if errParse != nil {
				log.Errorf("failed to parse object store endpoint %q: %v", env.objectEndpoint, errParse)
				return remoteStoreResult{}, errParse
			}
			switch strings.ToLower(parsed.Scheme) {
			case "http":
				useSSL = false
			case "https":
				useSSL = true
			default:
				errScheme := fmt.Errorf("unsupported object store scheme %q (only http and https are allowed)", parsed.Scheme)
				log.Errorf("%v", errScheme)
				return remoteStoreResult{}, errScheme
			}
			if parsed.Host == "" {
				errHost := fmt.Errorf("object store endpoint %q is missing host information", env.objectEndpoint)
				log.Errorf("%v", errHost)
				return remoteStoreResult{}, errHost
			}
			resolvedEndpoint = parsed.Host
			if parsed.Path != "" && parsed.Path != "/" {
				resolvedEndpoint = strings.TrimSuffix(parsed.Host+parsed.Path, "/")
			}
		}
		resolvedEndpoint = strings.TrimRight(resolvedEndpoint, "/")
		objCfg := store.ObjectStoreConfig{
			Endpoint:       resolvedEndpoint,
			Bucket:         env.objectBucket,
			AccessKey:      env.objectAccess,
			SecretKey:      env.objectSecret,
			LocalRoot:      objectStoreRoot,
			UseSSL:         useSSL,
			PathStyle:      true,
			OptionalConfig: isCloudDeploy,
		}
		objectStoreInst, err := store.NewObjectTokenStore(objCfg)
		if err != nil {
			log.Errorf("failed to initialize object token store: %v", err)
			return remoteStoreResult{}, err
		}
		examplePath := filepath.Join(wd, "config.example.yaml")
		cfg, err := objectStoreInst.BootstrapConfig(context.Background(), examplePath)
		if err != nil {
			log.Errorf("failed to bootstrap object-backed config: %v", err)
			return remoteStoreResult{}, err
		}
		cfg.AuthDir = objectStoreInst.AuthDir()
		log.Infof("object-backed token store enabled, bucket: %s", env.objectBucket)
		result.cfg = cfg
		result.configFilePath = objectStoreInst.ConfigPath()
		result.register = func() { sdkAuth.RegisterTokenStore(objectStoreInst) }
		return result, nil
	}
	if env.useGit {
		localPath := env.gitLocal
		if localPath == "" {
			if writableBase != "" {
				localPath = writableBase
			} else {
				localPath = wd
			}
		}
		gitStoreRoot := filepath.Join(localPath, "gitstore")
		authDir := filepath.Join(gitStoreRoot, "auths")
		gitStoreInst := store.NewGitTokenStore(env.gitURL, env.gitUser, env.gitPassword, env.gitBranch)
		gitStoreInst.SetBaseDir(authDir)
		gitStoreInst.SetOptionalConfig(isCloudDeploy)
		examplePath := filepath.Join(wd, "config.example.yaml")
		cfg, err := gitStoreInst.BootstrapConfig(context.Background(), examplePath)
		if err != nil {
			log.Errorf("failed to bootstrap git-backed config: %v", err)
			return remoteStoreResult{}, err
		}
		cfg.AuthDir = gitStoreInst.AuthDir()
		log.Infof("git-backed token store enabled, repository path: %s", gitStoreRoot)
		result.cfg = cfg
		result.configFilePath = gitStoreInst.ConfigPath()
		result.register = func() { sdkAuth.RegisterTokenStore(gitStoreInst) }
		return result, nil
	}
	return result, nil
}
