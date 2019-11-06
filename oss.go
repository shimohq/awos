package awos

import (
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/avast/retry-go"
	"io/ioutil"
	"strings"
	"time"
)

type OSS struct {
	Bucket *oss.Bucket
	Shards map[string]*oss.Bucket
}

func (ossClient *OSS) getBucket(key string) (*oss.Bucket, error) {
	if ossClient.Shards != nil && len(ossClient.Shards) > 0 {
		keyLength := len(key)
		bucket := ossClient.Shards[strings.ToLower(key[keyLength-1:keyLength])]
		if bucket == nil {
			return nil, errors.New("shards can't find bucket")
		}

		return bucket, nil
	}

	return ossClient.Bucket, nil
}

func (ossClient *OSS) Get(key string) (string, error) {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return "", err
	}

	body, err := bucket.GetObject(key)
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

func (ossClient *OSS) put(key string, data string, meta map[string]string, contentType string) error {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return err
	}

	options := make([]oss.Option, 0)
	if meta != nil {
		for k, v := range meta {
			options = append(options, oss.Meta(k, v))
		}
	}
	if contentType != "" {
		options = append(options, oss.ContentType(contentType))
	}

	return retry.Do(func() error {
		return bucket.PutObject(key, strings.NewReader(data), options...)
	}, retry.Attempts(3), retry.Delay(1*time.Second))
}

func (ossClient *OSS) Put(key string, data string, meta map[string]string) error {
	return ossClient.put(key, data, meta, "text/plain")
}

func (ossClient *OSS) PutWithContentType(key string, data string, meta map[string]string, contentType string) error {
	return ossClient.put(key, data, meta, contentType)
}

func (ossClient *OSS) Del(key string) error {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return err
	}

	return bucket.DeleteObject(key)
}

func (ossClient *OSS) Head(key string, attributes []string) (map[string]string, error) {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return nil, err
	}

	headers, err := bucket.GetObjectDetailedMeta(key)
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

func (ossClient *OSS) ListObject(key string, prefix string, marker string, maxKeys int, delimiter string) ([]string, error) {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return nil, err
	}

	res, err := bucket.ListObjects(oss.Prefix(prefix), oss.Marker(marker), oss.MaxKeys(maxKeys), oss.Delimiter(delimiter))
	keys := make([]string, 0)
	for _, v := range res.Objects {
		keys = append(keys, v.Key)
	}

	return keys, nil
}

func (ossClient *OSS) SignURL(key string, expired int64) (string, error) {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return "", err
	}

	return bucket.SignURL(key, oss.HTTPGet, expired)
}
