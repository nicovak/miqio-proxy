package r2

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/nicovak/miqio-go/logger"
	"go.uber.org/zap"
)

type Client struct {
	s3     *s3.Client
	bucket string
}

type Object struct {
	Body        io.ReadCloser
	ContentType string
	ETag        string
}

func NewClient(accountID, accessKeyID, secretAccessKey, bucket string) *Client {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	s3Client := s3.New(s3.Options{
		BaseEndpoint: aws.String(endpoint),
		Region:       "auto",
		Credentials:  credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
	})

	logger.Get().Infow("r2 client initialized",
		zap.String("bucket", bucket),
		zap.String("endpoint", endpoint),
	)

	return &Client{s3: s3Client, bucket: bucket}
}

func (c *Client) GetObject(ctx context.Context, key string) (*Object, error) {
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	contentType := "application/octet-stream"
	if out.ContentType != nil {
		contentType = *out.ContentType
	}

	etag := ""
	if out.ETag != nil {
		etag = *out.ETag
	}

	return &Object{
		Body:        out.Body,
		ContentType: contentType,
		ETag:        etag,
	}, nil
}

func (c *Client) HeadBucket(ctx context.Context) error {
	_, err := c.s3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})
	return err
}

func IsNotFound(err error) bool {
	var nsk *types.NoSuchKey
	if errors.As(err, &nsk) {
		return true
	}
	var nsb *types.NotFound
	return errors.As(err, &nsb)
}
