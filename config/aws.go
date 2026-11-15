package config

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

// AWSConfig overrides parts of the AWS SDK's default configuration. Anything left empty
// is resolved by the SDK's default chain: AWS_* environment variables, shared config and
// credentials files, SSO, web identity roles and instance roles.
type AWSConfig struct {
	Region  string `envconfig:"AWS_REGION"`
	Profile string `envconfig:"AWS_PROFILE"`
	// AccessKey and SecretKey are static credentials, used only when both are set.
	AccessKey string `envconfig:"AWS_ACCESS_KEY_ID"`
	SecretKey string `envconfig:"AWS_SECRET_ACCESS_KEY"`
}

// NewAWSConfig loads the AWS SDK configuration shared by every AWS client, with the overrides applied.
func NewAWSConfig(ctx context.Context, cfg AWSConfig) (aws.Config, error) {
	var opts []func(*awsconfig.LoadOptions) error
	if cfg.Region != "" {
		opts = append(opts, awsconfig.WithRegion(cfg.Region))
	}
	if cfg.Profile != "" {
		opts = append(opts, awsconfig.WithSharedConfigProfile(cfg.Profile))
	}
	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("failed to load aws config: %w", err)
	}
	return awsCfg, nil
}
