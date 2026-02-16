package util

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

const (
	envOSSRegion          = "OSS_REGION"
	envOSSBucketName      = "OSS_BUCKET_NAME"
	envOSSBucketURL       = "OSS_BUCKET_URL"
	envOSSEndpoint        = "OSS_ENDPOINT"
	envOSSAccessKeyID     = "OSS_ACCESS_KEY_ID"
	envOSSAccessKeySecret = "OSS_ACCESS_KEY_SECRET"

	objectNameRegex = `^[a-zA-Z0-9\-_\./]+$`
)

var validObjectNameRegex = regexp.MustCompile(objectNameRegex)

type OSSConfig struct {
	Region          string
	BucketName      string
	BucketURL       string
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
}

type OSSClient struct {
	oss        *oss.Client
	bucketName string
	bucketURL  string
	region     string
}

type PresignResult struct {
	URL           string
	SignedHeaders map[string]string
}

func NewOSSClient(cfg OSSConfig) (*OSSClient, error) {
	region, err := getConfigValue(cfg.Region, envOSSRegion, true)
	if err != nil {
		return nil, err
	}
	bucketName, err := getConfigValue(cfg.BucketName, envOSSBucketName, true)
	if err != nil {
		return nil, err
	}
	bucketURL, err := getConfigValue(cfg.BucketURL, envOSSBucketURL, false)
	if err != nil {
		return nil, err
	}
	endpoint, err := getConfigValue(cfg.Endpoint, envOSSEndpoint, false)
	if err != nil {
		return nil, err
	}

	if err := setCredentialsEnv(cfg.AccessKeyID, cfg.AccessKeySecret); err != nil {
		return nil, err
	}

	ossCfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(region).
		WithConnectTimeout(10 * time.Second).
		WithReadWriteTimeout(30 * time.Second).
		WithRetryMaxAttempts(3)
	if endpoint != "" {
		ossCfg = ossCfg.WithEndpoint(endpoint)
	}

	return &OSSClient{
		oss:        oss.NewClient(ossCfg),
		bucketName: bucketName,
		bucketURL:  bucketURL,
		region:     region,
	}, nil
}

func setCredentialsEnv(accessKeyID, accessKeySecret string) error {
	if strings.TrimSpace(accessKeyID) != "" {
		if err := os.Setenv(envOSSAccessKeyID, strings.TrimSpace(accessKeyID)); err != nil {
			return fmt.Errorf("failed to set OSS_ACCESS_KEY_ID: %w", err)
		}
	}
	if strings.TrimSpace(accessKeySecret) != "" {
		if err := os.Setenv(envOSSAccessKeySecret, strings.TrimSpace(accessKeySecret)); err != nil {
			return fmt.Errorf("failed to set OSS_ACCESS_KEY_SECRET: %w", err)
		}
	}
	return nil
}

func (c *OSSClient) PresignPut(ctx context.Context, objectName string, expires time.Duration) (*PresignResult, error) {
	if err := ValidateObjectName(objectName); err != nil {
		return nil, err
	}
	if expires <= 0 {
		return nil, errors.New("expiration must be positive")
	}

	request := &oss.PutObjectRequest{
		Bucket: oss.Ptr(c.bucketName),
		Key:    oss.Ptr(objectName),
	}
	result, err := c.oss.Presign(ctx, request, oss.PresignExpiration(time.Now().Add(expires)))
	if err != nil {
		return nil, fmt.Errorf("generate put presign URL failed (key: %s): %w", objectName, err)
	}

	return &PresignResult{
		URL:           result.URL,
		SignedHeaders: result.SignedHeaders,
	}, nil
}

func (c *OSSClient) PresignGet(ctx context.Context, objectName string, expires time.Duration) (*PresignResult, error) {
	if err := ValidateObjectName(objectName); err != nil {
		return nil, err
	}
	if expires <= 0 {
		return nil, errors.New("expiration must be positive")
	}

	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(c.bucketName),
		Key:    oss.Ptr(objectName),
	}
	result, err := c.oss.Presign(ctx, request, oss.PresignExpiration(time.Now().Add(expires)))
	if err != nil {
		return nil, fmt.Errorf("generate get presign URL failed (key: %s): %w", objectName, err)
	}

	return &PresignResult{
		URL:           result.URL,
		SignedHeaders: result.SignedHeaders,
	}, nil
}

func (c *OSSClient) BucketName() string {
	return c.bucketName
}

func (c *OSSClient) BucketURL() string {
	return c.bucketURL
}

func (c *OSSClient) Region() string {
	return c.region
}

func ValidateObjectName(objectName string) error {
	objectName = strings.TrimSpace(objectName)
	if objectName == "" {
		return errors.New("object name cannot be empty")
	}
	if strings.HasPrefix(objectName, "/") {
		return errors.New("object name cannot start with '/'")
	}
	if strings.Contains(objectName, "..") {
		return errors.New("object name cannot contain '..'")
	}
	if !validObjectNameRegex.MatchString(objectName) {
		return fmt.Errorf("invalid object name: %s (only letters, numbers, -, _, /, . are allowed)", objectName)
	}
	return nil
}

func getConfigValue(value string, envKey string, required bool) (string, error) {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value), nil
	}
	env := strings.TrimSpace(os.Getenv(envKey))
	if env == "" && required {
		return "", fmt.Errorf("environment variable %s is required", envKey)
	}
	return env, nil
}
