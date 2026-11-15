package config

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config configures access to a bucket on AWS S3 or an S3-compatible store such as
// MinIO, Cloudflare R2, Google Cloud Storage (HMAC keys), DigitalOcean Spaces or Backblaze B2.
type S3Config struct {
	Bucket string `envconfig:"S3_BUCKET"`
	// Endpoint is the URL of an S3-compatible store; empty means AWS S3.
	Endpoint string `envconfig:"AWS_ENDPOINT_URL_S3"`
	// Region overrides the AWS region for the bucket. S3-compatible stores still need one to
	// sign requests: "auto" for R2 and Google Cloud Storage, "us-east-1" for MinIO.
	Region string `envconfig:"S3_REGION"`
	// PublicEndpoint is the URL browsers reach the store at, used to sign object URLs when
	// the service reaches the store at an internal address. Defaults to Endpoint.
	PublicEndpoint string `envconfig:"S3_PUBLIC_ENDPOINT"`
}

// NewS3Client creates an S3 client for the configured store.
func NewS3Client(awsCfg aws.Config, cfg S3Config) *s3.Client {
	return s3.NewFromConfig(awsCfg, cfg.options(cfg.Endpoint))
}

// NewS3PresignClient creates a client that signs object URLs against the public endpoint.
// Signing happens offline, so the public endpoint never needs to be reachable from the service.
func NewS3PresignClient(awsCfg aws.Config, cfg S3Config) *s3.PresignClient {
	endpoint := cfg.PublicEndpoint
	if endpoint == "" {
		endpoint = cfg.Endpoint
	}
	return s3.NewPresignClient(s3.NewFromConfig(awsCfg, cfg.options(endpoint)))
}

func (c S3Config) options(endpoint string) func(*s3.Options) {
	return func(o *s3.Options) {
		if c.Region != "" {
			o.Region = c.Region
		}
		if endpoint == "" {
			return
		}
		o.BaseEndpoint = aws.String(endpoint)
		// S3-compatible stores address buckets by path, and several reject the
		// checksums the SDK sends by default.
		o.UsePathStyle = true
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	}
}
