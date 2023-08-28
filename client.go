package awos

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Client interface
type Client interface {
	Get(key string, options ...GetOptions) (string, error)
	GetBytes(key string, options ...GetOptions) ([]byte, error)
	GetAsReader(key string, options ...GetOptions) (io.ReadCloser, error)
	GetWithMeta(key string, attributes []string, options ...GetOptions) (io.ReadCloser, map[string]string, error)
	Put(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error
	Del(key string) error
	DelMulti(keys []string) error
	Head(key string, meta []string) (map[string]string, error)
	ListObject(key string, prefix string, marker string, maxKeys int, delimiter string) ([]string, error)
	SignURL(key string, expired int64, options ...SignOptions) (string, error)
	GetAndDecompress(key string) (string, error)
	GetAndDecompressAsReader(key string) (io.ReadCloser, error)
	CompressAndPut(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error
	Range(key string, offset int64, length int64) (io.ReadCloser, error)
	Exists(key string) (bool, error)
}

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
	// Only for s3-like, set http client timeout.
	// oss has default timeout, but s3 default timeout is 0 means no timeout.
	S3HttpTimeoutSecs              int64
	S3HttpTransportMaxConnsPerHost int
	S3HttpTransportIdleConnTimeout time.Duration
	// EnableCompressor
	EnableCompressor bool
	// CompressType gzip
	CompressType string
	// CompressLimit 大于该值之后才压缩 单位字节
	CompressLimit int
}

const (
	DefaultHttpTimeout = int64(60)
)

// New awos Client instance
func New(options *Options) (Client, error) {
	Register(DefaultGzipCompressor)
	storageType := strings.ToLower(options.StorageType)
	if storageType == StorageTypeOSS {
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
		if options.EnableCompressor {
			// 目前仅支持 gzip
			ossClient.cfg.EnableCompressor = options.EnableCompressor
			ossClient.cfg.CompressType = options.CompressType
			ossClient.cfg.CompressLimit = options.CompressLimit
			if comp, ok := compressors[options.CompressType]; ok {
				ossClient.compressor = comp
			} else {
				fmt.Printf("unknown type is: %s", options.CompressType)
			}
		}
		return ossClient, nil
	} else if storageType == StorageTypeS3 {
		var config *aws.Config

		// use minio
		if options.S3ForcePathStyle {
			config = &aws.Config{
				Region:           aws.String(options.Region),
				DisableSSL:       aws.Bool(!options.SSL),
				Credentials:      credentials.NewStaticCredentials(options.AccessKeyID, options.AccessKeySecret, ""),
				Endpoint:         aws.String(options.Endpoint),
				S3ForcePathStyle: aws.Bool(true),
			}
		} else {
			config = &aws.Config{
				Region:      aws.String(options.Region),
				DisableSSL:  aws.Bool(!options.SSL),
				Credentials: credentials.NewStaticCredentials(options.AccessKeyID, options.AccessKeySecret, ""),
			}
			if options.Endpoint != "" {
				config.Endpoint = aws.String(options.Endpoint)
			}
		}

		httpTimeout := DefaultHttpTimeout
		if options.S3HttpTimeoutSecs > 0 {
			httpTimeout = options.S3HttpTimeoutSecs
		}
		httpClient := &http.Client{
			Timeout: time.Second * time.Duration(httpTimeout),
		}
		if options.S3HttpTransportMaxConnsPerHost > 0 {
			transport := &http.Transport{
				MaxIdleConns:      options.S3HttpTransportMaxConnsPerHost,
				IdleConnTimeout:   30 * time.Second,
				MaxConnsPerHost:   options.S3HttpTransportMaxConnsPerHost,
				ForceAttemptHTTP2: true,
			}
			if options.S3HttpTransportIdleConnTimeout != 0 {
				transport.IdleConnTimeout = options.S3HttpTransportIdleConnTimeout
			}
			httpClient.Transport = transport
		}
		config.HTTPClient = httpClient
		service := s3.New(session.Must(session.NewSession(config)))

		var s3Client *S3
		if options.Shards != nil && len(options.Shards) > 0 {
			buckets := make(map[string]string)
			for _, v := range options.Shards {
				for i := 0; i < len(v); i++ {
					buckets[strings.ToLower(v[i:i+1])] = options.Bucket + "-" + v
				}
			}
			s3Client = &S3{
				ShardsBucket: buckets,
				Client:       service,
			}
		} else {
			s3Client = &S3{
				BucketName: options.Bucket,
				Client:     service,
			}
		}
		if options.EnableCompressor {
			// 目前仅支持 gzip
			if comp, ok := compressors[options.CompressType]; ok {
				s3Client.compressor = comp
			} else {
				fmt.Printf("unknown type is: %s", options.CompressType)
			}
		}
		return s3Client, nil
	} else {
		return nil, fmt.Errorf("Unknown StorageType:\"%s\", only supports oss,s3", options.StorageType)
	}
}
