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
)

type Client interface {
	Get(objectName string) (string, error)
	Put(key string, data string, meta map[string]string) error
	Del(key string) error
	Head(key string, meta []string) (map[string]string, error)
}

type Options struct {
	storage string
	oss     *OSSOptions
	aws     *AWSOptions
}

type OSSOptions struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	Bucket          string
}

type AWSOptions struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	SSL             bool
	Region          string
	Bucket          string
	// set S3ForcePathStyle = true when use minio
	S3ForcePathStyle bool
}

const (
	OSSStorage = "oss"
	S3Storage  = "aws"
)

func New(options *Options) (*Client, error) {
	var miossClient Client

	if options.storage == OSSStorage {
		ossConfig := options.oss
		client, err := oss.New(ossConfig.Endpoint, ossConfig.AccessKeyId, ossConfig.AccessKeySecret)
		if err != nil {
			return nil, err
		}

		bucket, err := client.Bucket(ossConfig.Bucket)
		if err != nil {
			return nil, err
		}

		ossClient := &osstorage.OSS{
			Bucket: bucket,
		}
		miossClient = ossClient

		return &miossClient, nil
	} else if options.storage == S3Storage {
		awsConfig := options.aws
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
		awsClient := &awss3.AWS{
			BucketName: awsConfig.Bucket,
			Client:     service,
		}
		miossClient = awsClient

		return &miossClient, nil
	} else {
		return nil, errors.New("options.storage should be oss or aws")
	}
}
