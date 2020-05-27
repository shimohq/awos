AWOS: Wrapper For Aliyun OSS And Amazon S3
====

[![](https://img.shields.io/badge/version-1.1.0-brightgreen.svg)](https://github.com/shimohq/awos)

awos for node:  https://github.com/shimohq/awos-js

## feat

- enable shards bucket
- add retry strategy
- avoid 404 status code:
    - `Get(objectName string) (string, error)` will return `"", nil` when object not exist
    - `Head(key string, meta []string) (map[string]string, error)` will return `nil, nil` when object not exist

## installing

Use go get to retrieve the SDK to add it to your GOPATH workspace, or project's Go module dependencies.

```
go get github.com/shimohq/awos
```

## how to use

### for Aliyun OSS

```golang
import "github.com/shimohq/awos"

ossClient, err := awos.New(&awos.Options{
    Storage: awos.OSSStorage,
    Oss: &awos.OSSOptions{
        AccessKeyId: "your accessKeyId",
        AccessKeySecret: "your accessKeySecret",
        Endpoint: "your endpoint",
        Bucket: "your bucket",
    },
})
```

### for Amazon S3

```golang
import "github.com/shimohq/awos"

awsClient, err := awos.New(&awos.Options{
    Storage: awos.S3Storage,
    Aws: &awos.AWSOptions{
        AccessKeyId: "your accessKeyId",
        AccessKeySecret: "your accessKeySecret",
        Bucket: "your bucket",
        // when use minio, S3ForcePathStyle must be set true
        // when use aws, endpoint is unnecessary and region must be set
        Region: "cn-north-1",
        Endpoint: "your endpoint",
        S3ForcePathStyle: true,
    },
})
```

the available operation：

```golang
Get(key string) (string, error)
GetAsReader(key string) (io.ReadCloser, error)
Put(key string, reader io.ReadSeeker, meta map[string]string, options ...PutOptions) error
Del(key string) error
Head(key string, meta []string) (map[string]string, error)
ListObject(key string, prefix string, marker string, maxKeys int, delimiter string) ([]string, error)
SignURL(key string, expired int64) (string, error)
```





