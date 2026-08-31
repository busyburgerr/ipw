// Package storage wraps an S3-compatible object store (MinIO in dev). Binary
// assets — avatars, portfolio images, project attachments — live here, never in
// Postgres.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"ipw/internal/config"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	mc            *minio.Client
	bucket        string
	publicBaseURL string
}

// New connects to the object store and ensures the bucket exists.
func New(ctx context.Context, cfg config.StorageConfig) (*Client, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init object store: %w", err)
	}

	exists, err := mc.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
		// Public read: avatars and portfolio images are served directly.
		// Private objects (contract deliverables) will use presigned URLs.
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/avatars/*", "arn:aws:s3:::%s/portfolio/*"]
			}]
		}`, cfg.Bucket, cfg.Bucket)
		if err := mc.SetBucketPolicy(ctx, cfg.Bucket, policy); err != nil {
			return nil, fmt.Errorf("set bucket policy: %w", err)
		}
	}

	return &Client{
		mc:            mc,
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}, nil
}

// Put stores data under prefix/<uuid><ext> and returns the object key.
func (c *Client) Put(ctx context.Context, prefix string, data []byte, contentType, ext string) (string, error) {
	key := path.Join(prefix, uuid.NewString()+ext)
	_, err := c.mc.PutObject(ctx, c.bucket, key, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}
	return key, nil
}

// Delete removes an object. A missing object is not an error.
func (c *Client) Delete(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}
	return c.mc.RemoveObject(ctx, c.bucket, key, minio.RemoveObjectOptions{})
}

// PublicURL returns the externally reachable URL for a stored key.
func (c *Client) PublicURL(key string) string {
	if key == "" {
		return ""
	}
	return c.publicBaseURL + "/" + key
}

// PresignedGet returns a time-limited URL for private objects (used later for
// contract deliverables).
func (c *Client) PresignedGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	u, err := c.mc.PresignedGetObject(ctx, c.bucket, key, ttl, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
