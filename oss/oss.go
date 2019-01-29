package oss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/avast/retry-go"
	"io/ioutil"
	"strings"
	"time"
)

type OSS struct {
	Bucket *oss.Bucket
}

func (ossClient *OSS) Get(objectName string) (string, error) {
	body, err := ossClient.Bucket.GetObject(objectName)
	defer func() {
		if body != nil {
			body.Close()
		}
	}()

	if err != nil {
		if oerr, ok := err.(oss.ServiceError); ok {
			if oerr.StatusCode == 404 {
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

func (ossClient *OSS) Put(key string, data string, meta map[string]string) error {
	options := make([]oss.Option, 0)
	if meta != nil {
		for k, v := range meta {
			options = append(options, oss.Meta(k, v))
		}
	}

	return retry.Do(func() error {
		return ossClient.Bucket.PutObject(key, strings.NewReader(data), options...)
	}, retry.Attempts(3), retry.Delay(1*time.Second))
}

func (ossClient *OSS) Del(key string) error {
	return ossClient.Bucket.DeleteObject(key)
}

func (ossClient *OSS) Head(key string, attributes []string) (map[string]string, error) {
	headers, err := ossClient.Bucket.GetObjectDetailedMeta(key)
	if err != nil {
		if oerr, ok := err.(oss.ServiceError); ok {
			if oerr.StatusCode == 404 {
				return nil, nil
			}
		}
		return nil, err
	}

	res := make(map[string]string)
	for _, v := range attributes {
		res[v] = headers.Get(v)
		if headers.Get(v) == "" {
			res[v] = headers.Get(oss.HTTPHeaderOssMetaPrefix + v)
		}
	}

	return res, nil
}
