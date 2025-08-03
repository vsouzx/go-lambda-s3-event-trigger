package service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	S3Client *s3.Client
}

func NewS3Service(s3Client *s3.Client) *S3Service {
	return &S3Service{
		S3Client: s3Client,
	}
}

func (s *S3Service) GetS3FileBytes(ctx context.Context, bucketName, keyName string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: &bucketName,
		Key:    &keyName,
	}

	result, err := s.S3Client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("erro ao baixar arquivo do S3: %w", err)
	}
	defer result.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, result.Body); err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo S3: %w", err)
	}

	return buf.Bytes(), nil
}
