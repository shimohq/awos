# AWOS: Wrapper For Aliyun OSS And Amazon S3

awos for node: https://github.com/shimohq/awos-js

## Features

- enable shards bucket
- add retry strategy
- avoid 404 status code:
  - `Get(objectName string) (string, error)` will return `"", nil` when object not exist
  - `Head(key string, meta []string) (map[string]string, error)` will return `nil, nil` when object not exist

## Installing

Use go get to retrieve the SDK to add it to your GOPATH workspace, or project's Go module dependencies.

```bash
go get github.com/shimohq/awos/v3
```

## How to use

```golang
import "github.com/shimohq/awos/v3"

awsClient, err := awos.New(&awos.Options{
    // Required, value is one of oss/s3, case insensetive
    StorageType: "string"
    // Required
    AccessKeyID: "string"
    // Required
    AccessKeySecret: "string"
    // Required
    Endpoint: "string"
    // Required
    Bucket: "string"
    // Optional, choose which bucket to use based on the last character of the key,
    // if bucket is 'content', shards is ['abc', 'edf'],
    // then the last character of the key with a/b/c will automatically use the content-abc bucket, and vice versa
    Shards: [2]string{"abc","def"}
    // Only for s3-like
    Region: "string"
    // Only for s3-like, whether to force path style URLs for S3 objects.
    S3ForcePathStyle: false
    // Only for s3-like
    SSL: false
})
```

Available operations：

```golang
Get(key string) (string, error)
GetAsReader(key string) (io.ReadCloser, error)
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
```
