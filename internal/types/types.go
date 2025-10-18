package types

// Analysis represents repository analysis results
type Analysis struct {
	RepoURL          string
	RepoPath         string
	AppDir           string // Subdirectory containing the main application code (relative to RepoPath)
	CommitSHA        string // Git commit SHA (if cloned from Git)
	Framework        string
	Language         string
	PackageManager   string // Package manager: "pip", "poetry", "uv", "pipenv", "npm", "yarn", etc.
	Dependencies     []string
	StartCommand     string
	Port             int
	EnvVars          map[string]string
	HasDockerfile    bool
	HasDockerCompose bool
	Verbose          bool // For detailed logging
}

// TerraformConfig represents generated Terraform configuration
type TerraformConfig struct {
	Path         string
	Directory    string
	Strategy     string
	AppName      string
	Region       string
	Framework    string
	Language     string
	Port         int
	RepoURL      string
	AppDir       string // Subdirectory containing the main application code
	StartCommand string
	EnvVars      map[string]string

	// EC2 sizing
	InstanceType string
	VolumeSize   int

	// Lambda sizing
	LambdaMemory              int
	LambdaTimeout             int
	LambdaReservedConcurrency int

	// EKS sizing
	EKSNodeType       string
	EKSMinNodes       int
	EKSMaxNodes       int
	EKSDesiredNodes   int
	EKSNodeVolumeSize int
}

// DeploymentResult represents deployment outcome
type DeploymentResult struct {
	Status        string
	Strategy      string
	Region        string
	Outputs       map[string]string
	TerraformDir  string
	Logs          []string
	Warnings      []string
	Optimizations []string
}

// DeploymentRule represents a heuristic decision rule
type DeploymentRule struct {
	Name           string
	Priority       int
	Description    string
	Conditions     RuleConditions
	Recommendation string
	InstanceType   string
	Reason         string
}

// RuleConditions defines conditions for a deployment rule
type RuleConditions struct {
	Framework        []string
	Language         string
	MinDependencies  int
	MaxDependencies  int
	HasDockerfile    *bool
	HasDockerCompose *bool
}

// DeploymentRules contains all deployment decision rules
type DeploymentRules struct {
	Version       string
	Rules         []DeploymentRule
	InstanceTypes map[string]InstanceTypeInfo
	Optimizations map[string]FrameworkOptimization
}

// InstanceTypeInfo contains EC2 instance type details
type InstanceTypeInfo struct {
	VCPU        int
	MemoryGB    int
	CostPerHour float64
	UseCases    []string
}

// FrameworkOptimization contains framework-specific deployment optimizations
type FrameworkOptimization struct {
	ProductionServer   string
	Workers            string
	RecommendedPorts   []int
	AdditionalPackages []string
	Notes              []string
}
