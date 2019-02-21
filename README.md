AWOS: Wrapper For OSS And AWS(MINIO)
====

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

### for oss

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

### for aws(minio)

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

### shards bucket（same usage for aws）

```golang
awsClient, _ := awos.New(&awos.Options{
    Storage: awos.OSSStorage,
    Oss: &awos.OSSOptions{
        AccessKeyId: "your accessKeyId",
        AccessKeySecret: "your accessKeySecret",
        Bucket: "your bucket",
        Shards: []string{"bucket-suffix1", "bucket-suffix2"},
        Endpoint: "your endpoint",
    },
})
```

the available operation：

```golang
Get(objectName string) (string, error)
Put(key string, data string, meta map[string]string) error
Del(key string) error
Head(key string, meta []string) (map[string]string, error)
```





