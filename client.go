package awos

import (
	"fmt"
	"io"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// Client interface
type Client interface {
	Get(key string, options ...GetOptions) (string, error)
	GetAsReader(key string, options ...GetOptions) (io.ReadCloser, error)
	GetWithMeta(key string, attributes []string, options ...GetOptions) (io.ReadCloser, map[string]string, error)
	Put(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error
	Del(key string) error
	DelMulti(keys []string) error
	Head(key string, meta []string) (map[string]string, error)
	ListObject(key string, prefix string, marker string, maxKeys int, delimiter string) ([]string, error)
	SignURL(key string, expired int64) (string, error)
	GetAndDecompress(key string) (string, error)
	GetAndDecompressAsReader(key string) (io.ReadCloser, error)
	CompressAndPut(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error
	Range(key string, offset int64, length int64) (io.ReadCloser, error)
}

type PresignEndpointReplacer func(endpoint string) string

type PresignedURLReplacer func(urlString string) string

// Options for New method
type Options struct {
	// Required, value is one of oss/s3, case insensetive
	StorageType string
	// Required
	AccessKeyID string
	// Required
	AccessKeySecret string
	// Required
	Endpoint string
	// Required
	Bucket string
	// Optional, choose which bucket to use based on the last character of the key,
	// if bucket is 'content', shards is ['abc', 'edf'],
	// then the last character of the key with a/b/c will automatically use the content-abc bucket, and vice versa
	Shards []string
	// Only for s3-like
	Region string
	// Only for s3-like, whether to force path style URLs for S3 objects.
	S3ForcePathStyle bool
	// Only for s3-like
	SSL bool
	// Only for s3-like, can only be either `v4` or `v2`
	//
	// Default: `v4`
	SignVersion string

	// Hooks
	//

	// Only for s3-like
	// This function will be executed before presigning
	ReplacePresignEndpoint PresignEndpointReplacer
	// Only for s3-like/oss
	// This function will be executed after presigned URL generated
	ReplacePresignedURL PresignedURLReplacer
}

func (o *Options) IsS3Like() bool {
	return o.S3ForcePathStyle
}

// New awos Client instance
func New(options *Options) (Client, error) {
	storageType := strings.ToLower(options.StorageType)

	if storageType == "oss" {
		client, err := oss.New(options.Endpoint, options.AccessKeyID, options.AccessKeySecret)
		if err != nil {
			return nil, err
		}

		var ossClient *OSS
		if options.Shards != nil && len(options.Shards) > 0 {
			buckets := make(map[string]*oss.Bucket)
			for _, v := range options.Shards {
				bucket, err := client.Bucket(options.Bucket + "-" + v)
				if err != nil {
					return nil, err
				}
				for i := 0; i < len(v); i++ {
					buckets[strings.ToLower(v[i:i+1])] = bucket
				}
			}

			ossClient = &OSS{
				Shards: buckets,
			}
		} else {
			bucket, err := client.Bucket(options.Bucket)
			if err != nil {
				return nil, err
			}

			ossClient = &OSS{
				Bucket: bucket,
			}
		}

		return ossClient, nil
	} else if storageType == "s3" {
		return newS3(options)
	} else {
		return nil, fmt.Errorf("Unknown StorageType:\"%s\", only supports oss,s3", options.StorageType)
	}
}
