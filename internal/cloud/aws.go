package cloud

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// AWSClient handles AWS operations
type AWSClient struct {
	ec2Client *ec2.Client
}

// NewAWSClient creates a new AWS client
func NewAWSClient(ctx context.Context) (*AWSClient, error) {
	// Load AWS config (uses default credential chain)
	// Use us-east-1 as default region for listing regions (the region doesn't matter for DescribeRegions)
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSClient{
		ec2Client: ec2.NewFromConfig(cfg),
	}, nil
}

// GetAllRegions returns all AWS regions
func (c *AWSClient) GetAllRegions(ctx context.Context) ([]string, error) {
	// Use DescribeRegions with AllRegions=true to get all regions including opt-in
	input := &ec2.DescribeRegionsInput{
		AllRegions: aws.Bool(true),
	}

	result, err := c.ec2Client.DescribeRegions(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe regions: %w", err)
	}

	regions := make([]string, 0, len(result.Regions))
	for _, region := range result.Regions {
		if region.RegionName != nil {
			regions = append(regions, *region.RegionName)
		}
	}

	// Sort alphabetically for better UX
	sort.Strings(regions)

	return regions, nil
}

// ValidateRegion checks if a region is valid
func (c *AWSClient) ValidateRegion(ctx context.Context, region string) (bool, error) {
	regions, err := c.GetAllRegions(ctx)
	if err != nil {
		return false, err
	}

	for _, r := range regions {
		if r == region {
			return true, nil
		}
	}

	return false, nil
}

// GetRegionForSelect returns regions formatted for selection (with descriptions)
func (c *AWSClient) GetRegionForSelect(ctx context.Context) ([]RegionOption, error) {
	regions, err := c.GetAllRegions(ctx)
	if err != nil {
		return nil, err
	}

	options := make([]RegionOption, 0, len(regions))
	for _, region := range regions {
		options = append(options, RegionOption{
			Code:        region,
			Description: getRegionDescription(region),
		})
	}

	return options, nil
}

// RegionOption represents a region with description
type RegionOption struct {
	Code        string
	Description string
}

// getRegionDescription returns a human-readable description for common regions
func getRegionDescription(region string) string {
	descriptions := map[string]string{
		"us-east-1":      "US East (N. Virginia)",
		"us-east-2":      "US East (Ohio)",
		"us-west-1":      "US West (N. California)",
		"us-west-2":      "US West (Oregon)",
		"eu-west-1":      "Europe (Ireland)",
		"eu-west-2":      "Europe (London)",
		"eu-west-3":      "Europe (Paris)",
		"eu-central-1":   "Europe (Frankfurt)",
		"eu-north-1":     "Europe (Stockholm)",
		"ap-northeast-1": "Asia Pacific (Tokyo)",
		"ap-northeast-2": "Asia Pacific (Seoul)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
		"ap-southeast-2": "Asia Pacific (Sydney)",
		"ap-south-1":     "Asia Pacific (Mumbai)",
		"ca-central-1":   "Canada (Central)",
		"sa-east-1":      "South America (São Paulo)",
	}

	if desc, ok := descriptions[region]; ok {
		return desc
	}

	return region // Fallback to region code
}
