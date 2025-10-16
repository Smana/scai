package types

// Analysis represents repository analysis results
type Analysis struct {
	RepoURL          string
	RepoPath         string
	Framework        string
	Language         string
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
	Path      string
	Directory string
	Strategy  string
}

// DeploymentResult represents deployment outcome
type DeploymentResult struct {
	Status       string
	PublicIP     string
	PublicURL    string
	TerraformDir string
	Logs         []string
	Strategy     string
	Warnings     []string
	Suggestions  []string
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
