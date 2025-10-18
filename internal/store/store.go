package store

import (
	"context"
	"time"

	"github.com/Smana/scia/internal/types"
)

// DeploymentStatus represents the current state of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending   DeploymentStatus = "pending"
	DeploymentStatusRunning   DeploymentStatus = "running"
	DeploymentStatusSucceeded DeploymentStatus = "succeeded"
	DeploymentStatusFailed    DeploymentStatus = "failed"
	DeploymentStatusDestroyed DeploymentStatus = "destroyed"
)

// Deployment represents a tracked deployment in the database
type Deployment struct {
	ID                string
	AppName           string
	UserPrompt        string
	RepoURL           string
	RepoCommitSHA     string
	Strategy          string
	Region            string
	Status            DeploymentStatus
	TerraformStateKey string
	TerraformDir      string

	// LLM information
	LLMProvider string
	LLMModel    string

	// Serialized as JSON
	Analysis      *types.Analysis
	Config        *types.TerraformConfig
	Outputs       map[string]string
	Warnings      []string
	Optimizations []string

	ErrorMessage string

	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeployedAt  *time.Time
	DestroyedAt *time.Time
}

// DeploymentFilter represents query filters for deployments
type DeploymentFilter struct {
	Region   string
	Strategy string
	Status   DeploymentStatus
	AppName  string
}

// Store defines the interface for deployment persistence
type Store interface {
	// Initialize creates tables and runs migrations
	Initialize(ctx context.Context) error

	// Close closes the database connection
	Close() error

	// Create creates a new deployment record
	Create(ctx context.Context, deployment *Deployment) error

	// Get retrieves a deployment by ID
	Get(ctx context.Context, id string) (*Deployment, error)

	// List retrieves all deployments with optional filtering
	List(ctx context.Context, filter *DeploymentFilter) ([]*Deployment, error)

	// Update updates a deployment record
	Update(ctx context.Context, deployment *Deployment) error

	// UpdateStatus updates only the status and error message
	UpdateStatus(ctx context.Context, id string, status DeploymentStatus, errorMessage string) error

	// Delete removes a deployment record
	Delete(ctx context.Context, id string) error
}
