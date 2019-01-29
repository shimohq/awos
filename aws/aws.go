package aws

import (
	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

type AWS struct {
	BucketName string
	Client     *s3.S3
}

func (a *AWS) Get(objectName string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(a.BucketName),
		Key:    aws.String(objectName),
	}

	result, err := a.Client.GetObject(input)

	body := result.Body

	defer func() {
		if body != nil {
			body.Close()
		}
	}()

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return "", nil
			}
		}
		return "", err
	}

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (a *AWS) Put(key string, data string, meta map[string]string) error {
	input := &s3.PutObjectInput{
		Body:   io.ReadSeeker(strings.NewReader(data)),
		Bucket: aws.String(a.BucketName),
		Key:    aws.String(key),
		Metadata: aws.StringMap(meta),
	}

	err := retry.Do(func() error {
		_, err := a.Client.PutObject(input)
		return err
	}, retry.Attempts(3), retry.Delay(1 * time.Second))

	return err
}

func (a *AWS) Del(key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(a.BucketName),
		Key:    aws.String(key),
	}

	_, err := a.Client.DeleteObject(input)
	return err
}

func (a *AWS) Head(key string, meta []string) (map[string]string, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(a.BucketName),
		Key:    aws.String(key),
	}

	result, err := a.Client.HeadObject(input)

	if err != nil {
		if aerr, ok := err.(awserr.RequestFailure); ok {
			if aerr.StatusCode() == 404 {
				return nil, nil
			}
		}
		return nil, err
	}

	// https://github.com/aws/aws-sdk-go/issues/445
	// aws 会将 meta 的首字母大写，在这里需要转换下
	res := make(map[string]string)
	for _, v := range meta {
		key := strings.Title(v)
		if result.Metadata[key] != nil {
			res[v] = *result.Metadata[key]
		}
	}

	return res, nil
}
