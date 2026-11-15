package config

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var testAWSConfig = aws.Config{
	Region:      "us-west-2",
	Credentials: credentials.NewStaticCredentialsProvider("test-key", "test-secret", ""),
}

func TestNewS3Client(t *testing.T) {
	type result struct {
		Region       string
		BaseEndpoint string
		UsePathStyle bool
		Request      aws.RequestChecksumCalculation
		Response     aws.ResponseChecksumValidation
	}
	tests := []struct {
		name     string
		cfg      S3Config
		expected result
	}{
		{
			name:     "aws s3",
			cfg:      S3Config{Bucket: "test-bucket"},
			expected: result{Region: "us-west-2"},
		},
		{
			name:     "aws s3 bucket region",
			cfg:      S3Config{Bucket: "test-bucket", Region: "eu-west-1"},
			expected: result{Region: "eu-west-1"},
		},
		{
			name: "s3-compatible store",
			cfg:  S3Config{Bucket: "test-bucket", Endpoint: "https://test-account.r2.cloudflarestorage.com", Region: "auto"},
			expected: result{
				Region:       "auto",
				BaseEndpoint: "https://test-account.r2.cloudflarestorage.com",
				UsePathStyle: true,
				Request:      aws.RequestChecksumCalculationWhenRequired,
				Response:     aws.ResponseChecksumValidationWhenRequired,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := NewS3Client(testAWSConfig, tt.cfg).Options()

			got := result{
				Region:       o.Region,
				BaseEndpoint: aws.ToString(o.BaseEndpoint),
				UsePathStyle: o.UsePathStyle,
				Request:      o.RequestChecksumCalculation,
				Response:     o.ResponseChecksumValidation,
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NewS3Client() options = %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestNewS3PresignClient(t *testing.T) {
	tests := []struct {
		name         string
		cfg          S3Config
		expectedHost string
		expectedPath string
	}{
		{
			name:         "aws s3",
			cfg:          S3Config{Bucket: "test-bucket"},
			expectedHost: "test-bucket.s3.us-west-2.amazonaws.com",
			expectedPath: "/test-key",
		},
		{
			name:         "endpoint",
			cfg:          S3Config{Bucket: "test-bucket", Endpoint: "http://minio:9000", Region: "us-east-1"},
			expectedHost: "minio:9000",
			expectedPath: "/test-bucket/test-key",
		},
		{
			name:         "public endpoint",
			cfg:          S3Config{Bucket: "test-bucket", Endpoint: "http://minio:9000", Region: "us-east-1", PublicEndpoint: "http://localhost:9000"},
			expectedHost: "localhost:9000",
			expectedPath: "/test-bucket/test-key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewS3PresignClient(testAWSConfig, tt.cfg).PresignGetObject(t.Context(), &s3.GetObjectInput{
				Bucket: aws.String(tt.cfg.Bucket),
				Key:    aws.String("test-key"),
			})
			if err != nil {
				t.Fatalf("PresignGetObject() error = %v", err)
			}

			u, err := url.Parse(req.URL)
			if err != nil {
				t.Fatalf("PresignGetObject() returned unparsable url %q: %v", req.URL, err)
			}
			if u.Host != tt.expectedHost || u.Path != tt.expectedPath {
				t.Errorf("PresignGetObject() = %s, want host %s and path %s", req.URL, tt.expectedHost, tt.expectedPath)
			}
		})
	}
}
