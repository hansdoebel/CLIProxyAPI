package main

// envLookup reads the first non-empty environment variable from keys.
type envLookup func(keys ...string) (string, bool)

// remoteStoreEnv is the process environment for optional Postgres, git, and object stores.
// Detection order matches the historical startup path, including git being cleared while
// Postgres is selected and then set again when GITSTORE_GIT_URL is also present.
// Opening still prefers Postgres, then object storage, then git.
type remoteStoreEnv struct {
	usePostgres    bool
	pgDSN          string
	pgSchema       string
	pgLocal        string
	useGit         bool
	gitURL         string
	gitUser        string
	gitPassword    string
	gitBranch      string
	gitLocal       string
	useObject      bool
	objectEndpoint string
	objectAccess   string
	objectSecret   string
	objectBucket   string
	objectLocal    string
}

func (e remoteStoreEnv) configured() bool {
	return e.usePostgres || e.useObject || e.useGit
}

func detectRemoteStoreEnv(lookup envLookup, writableBase, wd string) remoteStoreEnv {
	var env remoteStoreEnv
	if value, ok := lookup("PGSTORE_DSN", "pgstore_dsn"); ok {
		env.usePostgres = true
		env.pgDSN = value
	}
	if env.usePostgres {
		if value, ok := lookup("PGSTORE_SCHEMA", "pgstore_schema"); ok {
			env.pgSchema = value
		}
		if value, ok := lookup("PGSTORE_LOCAL_PATH", "pgstore_local_path"); ok {
			env.pgLocal = value
		}
		if env.pgLocal == "" {
			if writableBase != "" {
				env.pgLocal = writableBase
			} else {
				env.pgLocal = wd
			}
		}
		env.useGit = false
	}
	if value, ok := lookup("GITSTORE_GIT_URL", "gitstore_git_url"); ok {
		env.useGit = true
		env.gitURL = value
	}
	if value, ok := lookup("GITSTORE_GIT_USERNAME", "gitstore_git_username"); ok {
		env.gitUser = value
	}
	if value, ok := lookup("GITSTORE_GIT_TOKEN", "gitstore_git_token"); ok {
		env.gitPassword = value
	}
	if value, ok := lookup("GITSTORE_LOCAL_PATH", "gitstore_local_path"); ok {
		env.gitLocal = value
	}
	if value, ok := lookup("GITSTORE_GIT_BRANCH", "gitstore_git_branch"); ok {
		env.gitBranch = value
	}
	if value, ok := lookup("OBJECTSTORE_ENDPOINT", "objectstore_endpoint"); ok {
		env.useObject = true
		env.objectEndpoint = value
	}
	if value, ok := lookup("OBJECTSTORE_ACCESS_KEY", "objectstore_access_key"); ok {
		env.objectAccess = value
	}
	if value, ok := lookup("OBJECTSTORE_SECRET_KEY", "objectstore_secret_key"); ok {
		env.objectSecret = value
	}
	if value, ok := lookup("OBJECTSTORE_BUCKET", "objectstore_bucket"); ok {
		env.objectBucket = value
	}
	if value, ok := lookup("OBJECTSTORE_LOCAL_PATH", "objectstore_local_path"); ok {
		env.objectLocal = value
	}
	return env
}
