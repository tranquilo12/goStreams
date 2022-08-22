package subscriber

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func CreateS3Client() *s3.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("eu-central-1"),
	)
	if err != nil {
		panic(err)
	}

	s3Client := s3.NewFromConfig(cfg)
	return s3Client
}

// S3ListObjectsAPI defines the interface for the ListObjectsV2 function. Tests the function using a mocked service.
type S3ListObjectsAPI interface {
	ListObjectsV2(ctx context.Context,
		params *s3.ListObjectsV2Input,
		optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}
