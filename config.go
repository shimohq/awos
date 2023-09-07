package awos

type config struct {
	Debug bool
	bucketConfig
	Buckets   map[string]bucketConfig
	bucketKey string
}

type bucketConfig struct {
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
	S3HttpTimeoutSecs int64
	// EnableTraceInterceptor enable otel trace (only for s3)
	EnableTraceInterceptor bool
	// EnableMetricInterceptor enable prom metrics
	EnableMetricInterceptor bool
	// EnableClientTrace
	EnableClientTrace bool
	// EnableCompressor
	EnableCompressor bool
	// CompressType gzip
	CompressType string
	// CompressLimit 大于该值之后才压缩 单位字节
	CompressLimit int64
}

// DefaultConfig 返回默认配置
func DefaultConfig() *config {
	return &config{bucketConfig: bucketConfig{
		StorageType:             "s3",
		S3HttpTimeoutSecs:       60,
		EnableTraceInterceptor:  true,
		EnableMetricInterceptor: true,
	}}
}
