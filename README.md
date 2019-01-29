AWOS: Wrapper For OSS And AWS(MINIO)
====

## feat

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

```

```

### for aws(minio)

```

```

the available operation：

```
Get(objectName string) (string, error)
Put(key string, data string, meta map[string]string) error
Del(key string) error
Head(key string, meta []string) (map[string]string, error)
```





