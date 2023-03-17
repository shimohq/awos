package awos

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/golang/snappy"
)

var _ Client = (*S3)(nil)

type S3 struct {
	ShardsBucket map[string]string
	BucketName   string
	Client       *s3.S3
}

func (a *S3) getBucket(key string) (string, error) {
	if a.ShardsBucket != nil && len(a.ShardsBucket) > 0 {
		keyLength := len(key)
		bucketName := a.ShardsBucket[strings.ToLower(key[keyLength-1:keyLength])]
		if bucketName == "" {
			return "", errors.New("shards can't find bucket")
		}

		return bucketName, nil
	}

	return a.BucketName, nil
}

// don't forget to call the close() method of the io.ReadCloser
func (a *S3) GetAsReader(key string, options ...GetOptions) (io.ReadCloser, error) {
	bucketName, err := a.getBucket(key)
	if err != nil {
		return nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	setS3Options(options, input)

	result, err := a.Client.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return nil, nil
			}
		}
		return nil, err
	}

	return result.Body, err
}

// don't forget to call the close() method of the io.ReadCloser
func (a *S3) GetWithMeta(key string, attributes []string, options ...GetOptions) (io.ReadCloser, map[string]string, error) {
	bucketName, err := a.getBucket(key)
	if err != nil {
		return nil, nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	setS3Options(options, input)

	result, err := a.Client.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return nil, nil, nil
			}
		}
		return nil, nil, err
	}
	return result.Body, getS3Meta(attributes, mergeHttpStandardHeaders(&HeadGetObjectOutputWrapper{
		getObjectOutput: result,
	})), err
}

func (a *S3) Get(key string, options ...GetOptions) (string, error) {
	data, err := a.GetBytes(key, options...)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *S3) GetBytes(key string, options ...GetOptions) ([]byte, error) {
	result, err := a.get(key, options...)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	body := result.Body
	defer func() {
		if body != nil {
			body.Close()
		}
	}()

	return ioutil.ReadAll(body)
}

func (a *S3) Range(key string, offset int64, length int64) (io.ReadCloser, error) {
	readRange := fmt.Sprintf("bytes=%d-%d", offset, offset+length-1)
	input := &s3.GetObjectInput{
		Bucket: aws.String(a.BucketName),
		Key:    aws.String(key),
		Range:  &readRange,
	}
	r, err := a.Client.GetObject(input)
	if err != nil {
		return nil, err
	}
	return r.Body, nil
}

func (a *S3) GetAndDecompress(key string) (string, error) {
	result, err := a.get(key)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}

	body := result.Body
	defer func() {
		if body != nil {
			body.Close()
		}
	}()

	compressor := result.Metadata["Compressor"]
	if compressor != nil {
		if *compressor != "snappy" {
			return "", errors.New("GetAndDecompress only supports snappy for now, got " + *compressor)
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

		return string(decodedBytes), nil
	}

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (a *S3) GetAndDecompressAsReader(key string) (io.ReadCloser, error) {
	result, err := a.GetAndDecompress(key)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(strings.NewReader(result)), nil
}

func (a *S3) Put(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error {
	bucketName, err := a.getBucket(key)
	if err != nil {
		return err
	}

	putOptions := DefaultPutOptions()
	for _, opt := range options {
		opt(putOptions)
	}

	input := &s3.PutObjectInput{
		Body:        reader,
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Metadata:    aws.StringMap(meta),
		ContentType: aws.String(putOptions.contentType),
	}
	if putOptions.contentEncoding != nil {
		input.ContentEncoding = putOptions.contentEncoding
	}
	if putOptions.contentDisposition != nil {
		input.ContentDisposition = putOptions.contentDisposition
	}
	if putOptions.cacheControl != nil {
		input.CacheControl = putOptions.cacheControl
	}
	if putOptions.expires != nil {
		input.Expires = putOptions.expires
	}

	err = retry.Do(func() error {
		_, err := a.Client.PutObject(input)
		if err != nil && reader != nil {
			// Reset the body reader after the request since at this point it's already read
			// Note that it's safe to ignore the error here since the 0,0 position is always valid
			_, _ = reader.Seek(0, 0)
		}
		return err
	}, retry.Attempts(3), retry.Delay(1*time.Second))

	return err
}

func (a *S3) CompressAndPut(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	if meta == nil {
		meta = make(map[string]string)
	}

	encodedBytes := snappy.Encode(nil, data)

	meta["Compressor"] = "snappy"

	return a.Put(key, bytes.NewReader(encodedBytes), meta, options...)
}

func (a *S3) Del(key string) error {
	bucketName, err := a.getBucket(key)
	if err != nil {
		return err
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	_, err = a.Client.DeleteObject(input)
	return err
}

func (a *S3) DelMulti(keys []string) error {
	bucketsNameKeys := make(map[string][]string)
	for _, key := range keys {
		bucketName, err := a.getBucket(key)
		if err != nil {
			return err
		}
		bucketsNameKeys[bucketName] = append(bucketsNameKeys[bucketName], key)
	}

	for bucketName, BKeys := range bucketsNameKeys {
		delObjects := make([]*s3.ObjectIdentifier, len(BKeys))

		for idx, key := range BKeys {
			delObjects[idx] = &s3.ObjectIdentifier{
				Key: aws.String(key),
			}
		}

		input := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3.Delete{
				Objects: delObjects,
				Quiet:   aws.Bool(false),
			},
		}

		_, err := a.Client.DeleteObjects(input)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *S3) Head(key string, attributes []string) (map[string]string, error) {
	bucketName, err := a.getBucket(key)
	if err != nil {
		return nil, err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
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
	return getS3Meta(attributes, mergeHttpStandardHeaders(&HeadGetObjectOutputWrapper{
		headObjectOutput: result,
	})), nil
}

func (a *S3) ListObject(key string, prefix string, marker string, maxKeys int, delimiter string) ([]string, error) {
	bucketName, err := a.getBucket(key)
	if err != nil {
		return nil, err
	}

	input := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}
	if marker != "" {
		input.Marker = aws.String(marker)
	}
	if maxKeys > 0 {
		input.MaxKeys = aws.Int64(int64(maxKeys))
	}
	if delimiter != "" {
		input.Delimiter = aws.String(delimiter)
	}

	result, err := a.Client.ListObjects(input)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0)
	for _, v := range result.Contents {
		keys = append(keys, *v.Key)
	}

	return keys, nil
}

func (a *S3) SignURL(key string, expired int64) (string, error) {
	bucketName, err := a.getBucket(key)
	if err != nil {
		return "", err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	req, _ := a.Client.GetObjectRequest(input)
	return req.Presign(time.Duration(expired) * time.Second)
}

func (a *S3) SignURLWithProcess(key string, expired int64, process string) (string, error) {
	panic("not implemented")
}

func (a *S3) Exists(key string) (bool, error) {
	bucketName, err := a.getBucket(key)
	if err != nil {
		return false, err
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	_, err = a.Client.HeadObject(input)
	if err == nil {
		return true, nil
	}

	if aerr, ok := err.(awserr.RequestFailure); ok {
		if aerr.StatusCode() == 404 {
			return false, nil
		}
	}
	return false, err
}

func (a *S3) get(key string, options ...GetOptions) (*s3.GetObjectOutput, error) {
	bucketName, err := a.getBucket(key)
	if err != nil {
		return nil, err
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	setS3Options(options, input)

	result, err := a.Client.GetObject(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchKey {
				return nil, nil
			}
		}
		return nil, err
	}

	return result, nil
}

func getS3Meta(attributes []string, metaData map[string]*string) map[string]string {
	// https://github.com/aws/aws-sdk-go/issues/445
	// aws 会将 meta 的首字母大写，在这里需要转换下
	res := make(map[string]string)
	for _, v := range attributes {
		key := strings.Title(v)
		if metaData[key] != nil {
			res[v] = *metaData[key]
		}
	}
	return res
}

func setS3Options(options []GetOptions, getObjectInput *s3.GetObjectInput) {
	getOpts := DefaultGetOptions()
	for _, opt := range options {
		opt(getOpts)
	}
	if getOpts.contentEncoding != nil {
		getObjectInput.ResponseContentEncoding = getOpts.contentEncoding
	}
	if getOpts.contentType != nil {
		getObjectInput.ResponseContentType = getOpts.contentType
	}
}
