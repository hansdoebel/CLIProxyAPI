package main

import "testing"

func TestDetectRemoteStoreEnvPreservesBackendPrecedenceInputs(t *testing.T) {
	env := map[string]string{
		"PGSTORE_DSN":            "postgres://db",
		"PGSTORE_SCHEMA":         "cpa",
		"GITSTORE_GIT_URL":       "https://example.test/repo.git",
		"OBJECTSTORE_ENDPOINT":   "https://s3.example.test",
		"OBJECTSTORE_BUCKET":     "bucket",
		"OBJECTSTORE_ACCESS_KEY": "access",
		"OBJECTSTORE_SECRET_KEY": "secret",
	}
	lookup := func(keys ...string) (string, bool) {
		for _, key := range keys {
			if value, ok := env[key]; ok {
				return value, true
			}
		}
		return "", false
	}
	got := detectRemoteStoreEnv(lookup, "/data", "/work")
	if !got.usePostgres || got.pgDSN != "postgres://db" || got.pgSchema != "cpa" || got.pgLocal != "/data" {
		t.Fatalf("postgres env = %+v", got)
	}
	if !got.useGit || got.gitURL != "https://example.test/repo.git" {
		t.Fatalf("git env was not recorded after postgres selection: %+v", got)
	}
	if !got.useObject || got.objectEndpoint != "https://s3.example.test" || got.objectBucket != "bucket" {
		t.Fatalf("object env = %+v", got)
	}
	if !got.configured() {
		t.Fatal("configured() = false")
	}
}
