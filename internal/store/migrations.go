package store

const (
	// SchemaVersion is the current database schema version
	SchemaVersion = 1

	// InitialSchema creates the deployments table
	InitialSchema = `
CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY,
    app_name TEXT NOT NULL,
    user_prompt TEXT,
    repo_url TEXT NOT NULL,
    repo_commit_sha TEXT,
    strategy TEXT NOT NULL,
    region TEXT NOT NULL,
    status TEXT NOT NULL,
    terraform_state_key TEXT NOT NULL,
    terraform_dir TEXT,
    llm_provider TEXT,
    llm_model TEXT,
    analysis_json TEXT,
    config_json TEXT,
    outputs_json TEXT,
    warnings_json TEXT,
    optimizations_json TEXT,
    error_message TEXT,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    deployed_at DATETIME,
    destroyed_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_deployments_status ON deployments(status);
CREATE INDEX IF NOT EXISTS idx_deployments_app_name ON deployments(app_name);
CREATE INDEX IF NOT EXISTS idx_deployments_region ON deployments(region);
CREATE INDEX IF NOT EXISTS idx_deployments_strategy ON deployments(strategy);
CREATE INDEX IF NOT EXISTS idx_deployments_created_at ON deployments(created_at DESC);

CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME NOT NULL
);
`

)

// Migrations is a list of schema migrations to apply in order
var Migrations = []string{
	InitialSchema,
}
