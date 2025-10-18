package cloud

import (
	"context"
	"fmt"
)

// GCPClient handles GCP operations (stub for future implementation)
type GCPClient struct {
	// Future: Add GCP client
}

// NewGCPClient creates a new GCP client (stub)
func NewGCPClient(ctx context.Context) (*GCPClient, error) {
	// TODO: Implement GCP client initialization
	return nil, fmt.Errorf("GCP support not yet implemented")
}

// GetAllRegions returns all GCP regions (stub)
func (c *GCPClient) GetAllRegions(ctx context.Context) ([]string, error) {
	// TODO: Implement GCP region listing
	return nil, fmt.Errorf("GCP support not yet implemented")
}

// ValidateRegion checks if a GCP region is valid (stub)
func (c *GCPClient) ValidateRegion(ctx context.Context, region string) (bool, error) {
	// TODO: Implement GCP region validation
	return false, fmt.Errorf("GCP support not yet implemented")
}
