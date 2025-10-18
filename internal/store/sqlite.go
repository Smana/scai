package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SQLiteStore implements the Store interface using SQLite
type SQLiteStore struct {
	db     *sql.DB
	dbPath string
}

// NewSQLiteStore creates a new SQLite store
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &SQLiteStore{
		db:     db,
		dbPath: dbPath,
	}

	return store, nil
}

// Initialize creates tables and runs migrations
func (s *SQLiteStore) Initialize(ctx context.Context) error {
	// Check current schema version
	currentVersion, err := s.getSchemaVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schema version: %w", err)
	}

	// Apply migrations
	for i := currentVersion; i < len(Migrations); i++ {
		if err := s.applyMigration(ctx, i, Migrations[i]); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", i, err)
		}
	}

	return nil
}

// getSchemaVersion returns the current schema version
func (s *SQLiteStore) getSchemaVersion(ctx context.Context) (int, error) {
	var version int
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM schema_version
	`).Scan(&version)

	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		// If table doesn't exist, version is 0
		return 0, nil
	}

	return version, nil
}

// applyMigration applies a single migration
func (s *SQLiteStore) applyMigration(ctx context.Context, version int, migration string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // Rollback is safe to ignore on defer

	// Execute migration
	if _, err := tx.ExecContext(ctx, migration); err != nil {
		return err
	}

	// Record migration
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO schema_version (version, applied_at) VALUES (?, ?)
	`, version+1, time.Now()); err != nil {
		return err
	}

	return tx.Commit()
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Create creates a new deployment record
func (s *SQLiteStore) Create(ctx context.Context, deployment *Deployment) error {
	// Serialize JSON fields
	analysisJSON, err := json.Marshal(deployment.Analysis)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	configJSON, err := json.Marshal(deployment.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	outputsJSON, err := json.Marshal(deployment.Outputs)
	if err != nil {
		return fmt.Errorf("failed to marshal outputs: %w", err)
	}

	warningsJSON, err := json.Marshal(deployment.Warnings)
	if err != nil {
		return fmt.Errorf("failed to marshal warnings: %w", err)
	}

	optimizationsJSON, err := json.Marshal(deployment.Optimizations)
	if err != nil {
		return fmt.Errorf("failed to marshal optimizations: %w", err)
	}

	// Insert deployment
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO deployments (
			id, app_name, user_prompt, repo_url, repo_commit_sha,
			strategy, region, status, terraform_state_key, terraform_dir,
			llm_provider, llm_model,
			analysis_json, config_json, outputs_json, warnings_json, optimizations_json,
			error_message, created_at, updated_at, deployed_at, destroyed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		deployment.ID,
		deployment.AppName,
		deployment.UserPrompt,
		deployment.RepoURL,
		deployment.RepoCommitSHA,
		deployment.Strategy,
		deployment.Region,
		deployment.Status,
		deployment.TerraformStateKey,
		deployment.TerraformDir,
		deployment.LLMProvider,
		deployment.LLMModel,
		analysisJSON,
		configJSON,
		outputsJSON,
		warningsJSON,
		optimizationsJSON,
		deployment.ErrorMessage,
		deployment.CreatedAt,
		deployment.UpdatedAt,
		deployment.DeployedAt,
		deployment.DestroyedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert deployment: %w", err)
	}

	return nil
}

// Get retrieves a deployment by ID
func (s *SQLiteStore) Get(ctx context.Context, id string) (*Deployment, error) {
	var deployment Deployment
	var analysisJSON, configJSON, outputsJSON, warningsJSON, optimizationsJSON []byte
	var llmProvider, llmModel sql.NullString

	err := s.db.QueryRowContext(ctx, `
		SELECT
			id, app_name, user_prompt, repo_url, repo_commit_sha,
			strategy, region, status, terraform_state_key, terraform_dir,
			llm_provider, llm_model,
			analysis_json, config_json, outputs_json, warnings_json, optimizations_json,
			error_message, created_at, updated_at, deployed_at, destroyed_at
		FROM deployments
		WHERE id = ?
	`, id).Scan(
		&deployment.ID,
		&deployment.AppName,
		&deployment.UserPrompt,
		&deployment.RepoURL,
		&deployment.RepoCommitSHA,
		&deployment.Strategy,
		&deployment.Region,
		&deployment.Status,
		&deployment.TerraformStateKey,
		&deployment.TerraformDir,
		&llmProvider,
		&llmModel,
		&analysisJSON,
		&configJSON,
		&outputsJSON,
		&warningsJSON,
		&optimizationsJSON,
		&deployment.ErrorMessage,
		&deployment.CreatedAt,
		&deployment.UpdatedAt,
		&deployment.DeployedAt,
		&deployment.DestroyedAt,
	)

	// Convert sql.NullString to string
	if llmProvider.Valid {
		deployment.LLMProvider = llmProvider.String
	}
	if llmModel.Valid {
		deployment.LLMModel = llmModel.String
	}

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("deployment not found: %s", id)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Deserialize JSON fields
	if err := json.Unmarshal(analysisJSON, &deployment.Analysis); err != nil {
		return nil, fmt.Errorf("failed to unmarshal analysis: %w", err)
	}

	if err := json.Unmarshal(configJSON, &deployment.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := json.Unmarshal(outputsJSON, &deployment.Outputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outputs: %w", err)
	}

	if err := json.Unmarshal(warningsJSON, &deployment.Warnings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal warnings: %w", err)
	}

	if err := json.Unmarshal(optimizationsJSON, &deployment.Optimizations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal optimizations: %w", err)
	}

	return &deployment, nil
}

// buildListQuery builds the SQL query and args for List operation
func buildListQuery(filter *DeploymentFilter) (query string, args []interface{}) {
	query = `
		SELECT
			id, app_name, user_prompt, repo_url, repo_commit_sha,
			strategy, region, status, terraform_state_key, terraform_dir,
			llm_provider, llm_model,
			analysis_json, config_json, outputs_json, warnings_json, optimizations_json,
			error_message, created_at, updated_at, deployed_at, destroyed_at
		FROM deployments
		WHERE 1=1
	`
	args = []interface{}{}

	if filter != nil {
		if filter.Region != "" {
			query += " AND region = ?"
			args = append(args, filter.Region)
		}
		if filter.Strategy != "" {
			query += " AND strategy = ?"
			args = append(args, filter.Strategy)
		}
		if filter.Status != "" {
			query += " AND status = ?"
			args = append(args, filter.Status)
		}
		if filter.AppName != "" {
			query += " AND app_name = ?"
			args = append(args, filter.AppName)
		}
	}

	query += " ORDER BY created_at DESC"
	return query, args
}

// List retrieves all deployments with optional filtering
func (s *SQLiteStore) List(ctx context.Context, filter *DeploymentFilter) ([]*Deployment, error) {
	query, args := buildListQuery(filter)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	deployments := []*Deployment{}

	for rows.Next() {
		deployment, err := s.scanDeployment(rows)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, deployment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deployments: %w", err)
	}

	return deployments, nil
}

// scanDeployment scans a single deployment row and deserializes JSON fields
func (s *SQLiteStore) scanDeployment(rows *sql.Rows) (*Deployment, error) {
	var deployment Deployment
	var analysisJSON, configJSON, outputsJSON, warningsJSON, optimizationsJSON []byte
	var llmProvider, llmModel sql.NullString

	err := rows.Scan(
		&deployment.ID,
		&deployment.AppName,
		&deployment.UserPrompt,
		&deployment.RepoURL,
		&deployment.RepoCommitSHA,
		&deployment.Strategy,
		&deployment.Region,
		&deployment.Status,
		&deployment.TerraformStateKey,
		&deployment.TerraformDir,
		&llmProvider,
		&llmModel,
		&analysisJSON,
		&configJSON,
		&outputsJSON,
		&warningsJSON,
		&optimizationsJSON,
		&deployment.ErrorMessage,
		&deployment.CreatedAt,
		&deployment.UpdatedAt,
		&deployment.DeployedAt,
		&deployment.DestroyedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan deployment: %w", err)
	}

	// Convert sql.NullString to string
	if llmProvider.Valid {
		deployment.LLMProvider = llmProvider.String
	}
	if llmModel.Valid {
		deployment.LLMModel = llmModel.String
	}

	// Deserialize JSON fields
	if err := s.deserializeJSONFields(&deployment, analysisJSON, configJSON, outputsJSON, warningsJSON, optimizationsJSON); err != nil {
		return nil, err
	}

	return &deployment, nil
}

// deserializeJSONFields unmarshals JSON data into deployment fields
func (s *SQLiteStore) deserializeJSONFields(deployment *Deployment, analysisJSON, configJSON, outputsJSON, warningsJSON, optimizationsJSON []byte) error {
	if err := json.Unmarshal(analysisJSON, &deployment.Analysis); err != nil {
		return fmt.Errorf("failed to unmarshal analysis: %w", err)
	}
	if err := json.Unmarshal(configJSON, &deployment.Config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if err := json.Unmarshal(outputsJSON, &deployment.Outputs); err != nil {
		return fmt.Errorf("failed to unmarshal outputs: %w", err)
	}
	if err := json.Unmarshal(warningsJSON, &deployment.Warnings); err != nil {
		return fmt.Errorf("failed to unmarshal warnings: %w", err)
	}
	if err := json.Unmarshal(optimizationsJSON, &deployment.Optimizations); err != nil {
		return fmt.Errorf("failed to unmarshal optimizations: %w", err)
	}
	return nil
}

// Update updates a deployment record
func (s *SQLiteStore) Update(ctx context.Context, deployment *Deployment) error {
	deployment.UpdatedAt = time.Now()

	// Serialize JSON fields
	analysisJSON, err := json.Marshal(deployment.Analysis)
	if err != nil {
		return fmt.Errorf("failed to marshal analysis: %w", err)
	}

	configJSON, err := json.Marshal(deployment.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	outputsJSON, err := json.Marshal(deployment.Outputs)
	if err != nil {
		return fmt.Errorf("failed to marshal outputs: %w", err)
	}

	warningsJSON, err := json.Marshal(deployment.Warnings)
	if err != nil {
		return fmt.Errorf("failed to marshal warnings: %w", err)
	}

	optimizationsJSON, err := json.Marshal(deployment.Optimizations)
	if err != nil {
		return fmt.Errorf("failed to marshal optimizations: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE deployments SET
			app_name = ?,
			user_prompt = ?,
			repo_url = ?,
			repo_commit_sha = ?,
			strategy = ?,
			region = ?,
			status = ?,
			terraform_state_key = ?,
			terraform_dir = ?,
			llm_provider = ?,
			llm_model = ?,
			analysis_json = ?,
			config_json = ?,
			outputs_json = ?,
			warnings_json = ?,
			optimizations_json = ?,
			error_message = ?,
			updated_at = ?,
			deployed_at = ?,
			destroyed_at = ?
		WHERE id = ?
	`,
		deployment.AppName,
		deployment.UserPrompt,
		deployment.RepoURL,
		deployment.RepoCommitSHA,
		deployment.Strategy,
		deployment.Region,
		deployment.Status,
		deployment.TerraformStateKey,
		deployment.TerraformDir,
		deployment.LLMProvider,
		deployment.LLMModel,
		analysisJSON,
		configJSON,
		outputsJSON,
		warningsJSON,
		optimizationsJSON,
		deployment.ErrorMessage,
		deployment.UpdatedAt,
		deployment.DeployedAt,
		deployment.DestroyedAt,
		deployment.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	return nil
}

// UpdateStatus updates only the status and error message
func (s *SQLiteStore) UpdateStatus(ctx context.Context, id string, status DeploymentStatus, errorMessage string) error {
	var deployedAt *time.Time
	var destroyedAt *time.Time
	now := time.Now()

	if status == DeploymentStatusSucceeded {
		deployedAt = &now
	} else if status == DeploymentStatusDestroyed {
		destroyedAt = &now
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE deployments SET
			status = ?,
			error_message = ?,
			updated_at = ?,
			deployed_at = COALESCE(deployed_at, ?),
			destroyed_at = COALESCE(destroyed_at, ?)
		WHERE id = ?
	`, status, errorMessage, now, deployedAt, destroyedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update deployment status: %w", err)
	}

	return nil
}

// Delete removes a deployment record
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM deployments WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}
	return nil
}
