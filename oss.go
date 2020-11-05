package awos

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/avast/retry-go"
	"github.com/golang/snappy"
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

// don't forget to call the close() method of the io.ReadCloser
func (ossClient *OSS) GetAsReader(key string) (io.ReadCloser, error) {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return nil, err
	}

	return bucket.GetObject(key)
}

func (ossClient *OSS) Get(key string) (string, error) {
	result, err := ossClient.get(key)

	if err != nil {
		return "", err
	}

	if result == nil {
		return "", nil
	}

	body := result.Response
	defer func() {
		if body != nil {
			body.Close()
		}
	}()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (ossClient *OSS) GetAndDecompress(key string) (string, error) {
	result, err := ossClient.get(key)

	if err != nil {
		return "", err
	}

	if result == nil {
		return "", nil
	}

	body := result.Response
	defer func() {
		if body != nil {
			body.Close()
		}
	}()

	compressor := body.Headers.Get("X-Oss-Meta-Compressor")
	if compressor != "" {
		if compressor != "snappy" {
			return "", errors.New("GetAndDecompress only supports snappy for now, got " + compressor)
		}

		rawBytes, err := ioutil.ReadAll(body)
		if err != nil {
			return "", err
		}

		decodedBytes, err := snappy.Decode(nil, rawBytes)
		if err != nil {
			if errors.Is(err, snappy.ErrCorrupt) {
				reader := snappy.NewReader(bytes.NewReader(rawBytes))
				data, err := ioutil.ReadAll(reader)
				if err != nil {
					return "", err
				}

				return string(data), nil
			}
			return "", err
		}

		return string(decodedBytes), err
	}

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (ossClient *OSS) GetAndDecompressAsReader(key string) (io.ReadCloser, error) {
	ret, err := ossClient.GetAndDecompress(key)
	if err != nil {
		return nil, err
	}
	return ioutil.NopCloser(strings.NewReader(ret)), nil
}

func (ossClient *OSS) get(key string) (*oss.GetObjectResult, error) {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return nil, err
	}

	result, err := bucket.DoGetObject(&oss.GetObjectRequest{ObjectKey: key}, []oss.Option{})

	if err != nil {
		if oerr, ok := err.(oss.ServiceError); ok {
			if oerr.StatusCode == 404 {
				return nil, nil
			}
		}
		return nil, err
	}

	return result, nil
}

func (ossClient *OSS) Put(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return err
	}

	putOptions := DefaultPutOptions
	for _, opt := range options {
		opt(putOptions)
	}

	ossOptions := make([]oss.Option, 0)
	if meta != nil {
		for k, v := range meta {
			ossOptions = append(ossOptions, oss.Meta(k, v))
		}
	}
	ossOptions = append(ossOptions, oss.ContentType(putOptions.contentType))

	return retry.Do(func() error {
		err := bucket.PutObject(key, reader, ossOptions...)
		if err != nil && reader != nil {
			// Reset the body reader after the request since at this point it's already read
			// Note that it's safe to ignore the error here since the 0,0 position is always valid
			_, _ = reader.Seek(0, 0)
		}
		return err
	}, retry.Attempts(3), retry.Delay(1*time.Second))
}

func (ossClient *OSS) CompressAndPut(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	if meta == nil {
		meta = make(map[string]string)
	}

	encodedBytes := snappy.Encode(nil, data)

	meta["Compressor"] = "snappy"

	return ossClient.Put(key, bytes.NewReader(encodedBytes), meta, options...)
}

func (ossClient *OSS) Del(key string) error {
	bucket, err := ossClient.getBucket(key)
	if err != nil {
		return err
	}

	return bucket.DeleteObject(key)
}

func (ossClient *OSS) DelMulti(keys []string) error {
	_, err := ossClient.Bucket.DeleteObjects(keys)

	return err
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
