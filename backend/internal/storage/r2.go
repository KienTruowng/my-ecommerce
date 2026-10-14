package storage

import (
	"context"
	"fmt"

	appconfig "github.com/AiEmStylix/ecommerce/backend/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

type R2Storage struct {
	Client *s3.Client
	PresignClient *s3.PresignClient
}

func NewR2Storage(cfg *appconfig.Config) (*R2Storage, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.R2AccessKey, cfg.R2SecretKey, "")

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion("auto"), awsconfig.WithCredentialsProvider(creds))

	if err != nil {
		return nil, fmt.Errorf("không thể khởi tạo cấu hình AWS/R2: %w", err)
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID)

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &R2Storage{
		Client:        client,
		PresignClient: s3.NewPresignClient(client),
	}, nil
}