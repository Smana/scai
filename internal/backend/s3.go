package backend

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	// DefaultAWSRegion is the AWS default region (legacy behavior for S3)
	DefaultAWSRegion = "us-east-1"
)

// S3Manager handles S3 operations for Terraform state backend
type S3Manager struct {
	client *s3.Client
	region string
}

// NewS3Manager creates a new S3 manager
func NewS3Manager(ctx context.Context, region string) (*S3Manager, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &S3Manager{
		client: s3.NewFromConfig(cfg),
		region: region,
	}, nil
}

// BucketExists checks if an S3 bucket exists
func (m *S3Manager) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	_, err := m.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// Check if it's a "not found" error
		return false, nil
	}

	return true, nil
}

// ListBuckets returns all S3 buckets in the account
func (m *S3Manager) ListBuckets(ctx context.Context) ([]string, error) {
	result, err := m.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	buckets := make([]string, 0, len(result.Buckets))
	for _, bucket := range result.Buckets {
		if bucket.Name != nil {
			buckets = append(buckets, *bucket.Name)
		}
	}

	return buckets, nil
}

// GetBucketLocation returns the AWS region where a bucket is located
func (m *S3Manager) GetBucketLocation(ctx context.Context, bucketName string) (string, error) {
	result, err := m.client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get bucket location: %w", err)
	}

	// AWS returns empty string for us-east-1 (legacy behavior)
	if result.LocationConstraint == "" {
		return DefaultAWSRegion, nil
	}

	return string(result.LocationConstraint), nil
}

// CreateStateBucket creates and configures an S3 bucket for Terraform state
// Returns true if the bucket was created, false if it already existed
func (m *S3Manager) CreateStateBucket(ctx context.Context, bucketName string) (bool, error) {
	// Check if bucket already exists
	exists, err := m.BucketExists(ctx, bucketName)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	bucketCreated := false
	// Step 1: Create the bucket (skip if already exists)
	if !exists {
		createInput := &s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		}

		// For regions other than us-east-1, we need to specify location constraint
		if m.region != DefaultAWSRegion {
			createInput.CreateBucketConfiguration = &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraint(m.region),
			}
		}

		_, err := m.client.CreateBucket(ctx, createInput)
		if err != nil {
			return false, fmt.Errorf("failed to create bucket: %w", err)
		}
		bucketCreated = true
	}

	// Step 2: Enable versioning for state recovery
	_, err = m.client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to enable versioning: %w", err)
	}

	// Step 3: Enable server-side encryption (AES256)
	_, err = m.client.PutBucketEncryption(ctx, &s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucketName),
		ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
			Rules: []types.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
						SSEAlgorithm: types.ServerSideEncryptionAes256,
					},
					BucketKeyEnabled: aws.Bool(true),
				},
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to enable encryption: %w", err)
	}

	// Step 4: Block all public access
	_, err = m.client.PutPublicAccessBlock(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
		PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to block public access: %w", err)
	}

	// Step 5: Add lifecycle policy to cleanup old lock files
	_, err = m.client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
		LifecycleConfiguration: &types.BucketLifecycleConfiguration{
			Rules: []types.LifecycleRule{
				{
					ID:     aws.String("cleanup-lock-files"),
					Status: types.ExpirationStatusEnabled,
					Filter: &types.LifecycleRuleFilter{
						Prefix: aws.String(".terraform.tfstate.lock.info"),
					},
					NoncurrentVersionExpiration: &types.NoncurrentVersionExpiration{
						NoncurrentDays: aws.Int32(7),
					},
				},
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to set lifecycle policy: %w", err)
	}

	// Step 6: Add tags
	_, err = m.client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
		Bucket: aws.String(bucketName),
		Tagging: &types.Tagging{
			TagSet: []types.Tag{
				{
					Key:   aws.String("ManagedBy"),
					Value: aws.String("SCAI"),
				},
				{
					Key:   aws.String("Purpose"),
					Value: aws.String("Terraform State"),
				},
			},
		},
	})
	if err != nil {
		return false, fmt.Errorf("failed to add tags: %w", err)
	}

	return bucketCreated, nil
}
