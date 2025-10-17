package ui

// DeploymentPlan represents the complete deployment plan
type DeploymentPlan struct {
	Strategy  string
	Region    string
	AppName   string
	Resources []ResourceConfig
}

// ResourceConfig represents a single resource to be created
type ResourceConfig struct {
	Type       string            // Resource type (e.g., "VPC", "EC2 Instance", "EKS Cluster")
	Name       string            // Resource name
	Parameters map[string]string // Configuration parameters
	Important  bool              // Highlight important resources
}

// Add a parameter to a resource
func (r *ResourceConfig) AddParameter(key, value string) {
	if r.Parameters == nil {
		r.Parameters = make(map[string]string)
	}
	r.Parameters[key] = value
}
