package media

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/AiEmStylix/ecommerce/backend/internal/config"
	"github.com/AiEmStylix/ecommerce/backend/internal/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MediaService struct {
	r2 *storage.R2Storage
	cfg *config.Config
}

func NewMediaService(r2 *storage.R2Storage, cfg *config.Config) *MediaService {
	return &MediaService{r2:r2, cfg: cfg}
}

func (s *MediaService) GenerateUploadURL (folder, filename, filetype string) (string, string, error) {
	objectKey := fmt.Sprintf("%s/%d_%s", folder, time.Now().Unix(), filepath.Base(filename))

	req, err := s.r2.PresignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(s.cfg.R2Bucket),
		Key: aws.String(objectKey),
		ContentType: aws.String(filetype),
	}, s3.WithPresignExpires(5*time.Minute))

	if err != nil {
		return "", "", errors.New("không thể tạo link upload")
	}

	return req.URL, objectKey, nil
}

func (s*MediaService) VerifyUploadFile(objectKey string, maxSizeBytes int64) (string, error) {
	finalURL := fmt.Sprintf("%s/%s", s.cfg.R2PublicURL, objectKey)
	
	resp, err := http.Head(finalURL)

	if err != nil {
		return "", errors.New("không thể kết nối đến máy chủ lưu trữ")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("file chưa được tải lên thành công hoặc không tồn tại")
	}

	if resp.ContentLength > maxSizeBytes {
		return "", fmt.Errorf("kích thước file vượt quá giới hạn (%d bytes)", maxSizeBytes)
	}

	return finalURL, nil
}