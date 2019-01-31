package awos

import (
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	awss3 "github.com/shimohq/awos/aws"
	osstorage "github.com/shimohq/awos/oss"
	"strings"
)

type Client interface {
	Get(objectName string) (string, error)
	Put(key string, data string, meta map[string]string) error
	Del(key string) error
	Head(key string, meta []string) (map[string]string, error)
}

type Options struct {
	Storage string
	Oss     *OSSOptions
	Aws     *AWSOptions
}

type OSSOptions struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	Bucket          string
	Shards          []string
}

type AWSOptions struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	SSL             bool
	Region          string
	Bucket          string
	Shards          []string
	// set S3ForcePathStyle = true when use minio
	S3ForcePathStyle bool
}

const (
	OSSStorage = "oss"
	S3Storage  = "aws"
)

func New(options *Options) (Client, error) {
	var miossClient Client

	if options.Storage == OSSStorage {
		ossConfig := options.Oss
		client, err := oss.New(ossConfig.Endpoint, ossConfig.AccessKeyId, ossConfig.AccessKeySecret)
		if err != nil {
			return nil, err
		}

		var ossClient *osstorage.OSS
		if ossConfig.Shards != nil && len(ossConfig.Shards) > 0 {
			buckets := make(map[string]*oss.Bucket)
			for _, v := range ossConfig.Shards {
				bucket, err := client.Bucket(ossConfig.Bucket + "-" + v)
				if err != nil {
					return nil, err
				}
				for i := 0; i < len(v); i++ {
					buckets[strings.ToLower(v[i:i+1])] = bucket
				}
			}

			ossClient = &osstorage.OSS{
				Shards: buckets,
			}
		} else {
			bucket, err := client.Bucket(ossConfig.Bucket)
			if err != nil {
				return nil, err
			}

			ossClient = &osstorage.OSS{
				Bucket: bucket,
			}
		}

		miossClient = ossClient

		return miossClient, nil
	} else if options.Storage == S3Storage {
		awsConfig := options.Aws
		var sess *session.Session

		// use minio
		if awsConfig.S3ForcePathStyle == true {
			sess = session.Must(session.NewSession(&aws.Config{
				Region:           aws.String(awsConfig.Region),
				DisableSSL:       aws.Bool(awsConfig.SSL == false),
				Credentials:      credentials.NewStaticCredentials(awsConfig.AccessKeyId, awsConfig.AccessKeySecret, ""),
				Endpoint:         aws.String(awsConfig.Endpoint),
				S3ForcePathStyle: aws.Bool(true),
			}))
		} else {
			sess = session.Must(session.NewSession(&aws.Config{
				Region:      aws.String(awsConfig.Region),
				DisableSSL:  aws.Bool(awsConfig.SSL == false),
				Credentials: credentials.NewStaticCredentials(awsConfig.AccessKeyId, awsConfig.AccessKeySecret, ""),
			}))
		}
		service := s3.New(sess)

		var awsClient *awss3.AWS
		if awsConfig.Shards != nil && len(awsConfig.Shards) > 0 {
			buckets := make(map[string]string)
			for _, v := range awsConfig.Shards {
				for i := 0; i < len(v); i++ {
					buckets[strings.ToLower(v[i:i+1])] = awsConfig.Bucket + "-" + v
				}
			}
			awsClient = &awss3.AWS{
				ShardsBucket: buckets,
				Client:       service,
			}
		} else {
			awsClient = &awss3.AWS{
				BucketName: awsConfig.Bucket,
				Client:     service,
			}
		}

		miossClient = awsClient

		return miossClient, nil
	} else {
		return nil, errors.New("options.storage should be oss or aws")
	}
}
