package awos

/**
AccessKeyId=${accessKeyId} AccessKeySecret=${accessKeySecret} bucket=${bucket} Endpoint=${endpoint} go test -v oss_test.go oss.go
*/

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"os"
	"strconv"
	"testing"
)

const (
	guid         = "test123"
	content      = "aaaaaa"
	expectLength = 6
	expectHead   = 1
)

var (
	ossClient *OSS
)

func init() {
	client, err := oss.New(os.Getenv("Endpoint"), os.Getenv("AccessKeyId"),
		os.Getenv("AccessKeySecret"))
	if err != nil {
		panic(err)
	}

	bucket, err := client.Bucket(os.Getenv("bucket"))
	if err != nil {
		panic(err)
	}

	ossClient = &OSS{
		Bucket: bucket,
	}
}

func TestOSS_Put(t *testing.T) {
	meta := make(map[string]string)
	meta["head"] = strconv.Itoa(expectHead)
	meta["length"] = strconv.Itoa(expectLength)

	err := ossClient.Put(guid, content, meta)
	if err != nil {
		t.Log("oss put error", err)
		t.Fail()
	}
}

func TestOSS_Head(t *testing.T) {
	attributes := make([]string, 0)
	attributes = append(attributes, "head")
	var res map[string]string
	var err error
	var head int
	var length int

	res, err = ossClient.Head(guid, attributes)
	if err != nil {
		t.Log("oss head error", err)
		t.Fail()
	}

	head, err = strconv.Atoi(res["head"])
	if err != nil || head != 1 {
		t.Log("oss get head fail, res:", res, "err:", err)
		t.Fail()
	}

	attributes = append(attributes, "length")
	attributes = append(attributes, "Content-Type")
	res, err = ossClient.Head(guid, attributes)
	if err != nil {
		t.Log("oss head error", err)
		t.Fail()
	}

	head, err = strconv.Atoi(res["head"])
	length, err = strconv.Atoi(res["length"])
	if err != nil || head != expectHead || length != expectLength {
		t.Log("oss get head fail, res:", res, "err:", err)
		t.Fail()
	}

	if res["Content-Type"] != "text/plain" {
		t.Log("oss get head Content-Type wrong, res:", res, "err:", err)
		t.Fail()
	}
}

func TestOSS_Get(t *testing.T) {
	res, err := ossClient.Get(guid)
	if err != nil || res != content {
		t.Log("oss get content fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestOSS_SignURL(t *testing.T) {
	res, err := ossClient.SignURL(guid, 60)
	if err != nil {
		t.Log("oss signUrl fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestOSS_ListObject(t *testing.T) {
	res, err := ossClient.ListObject(guid, guid[0:4], "", 10, "")
	if err != nil || len(res) == 0 {
		t.Log("oss list objects fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestOSS_Del(t *testing.T) {
	err := ossClient.Del(guid)
	if err != nil {
		t.Log("oss del key fail, err:", err)
		t.Fail()
	}
}

func TestOSS_GetNotExist(t *testing.T) {
	res1, err := ossClient.Get(guid + "123")
	if res1 != "" || err != nil {
		t.Log("oss get not exist key fail, res:", res1, "err:", err)
		t.Fail()
	}

	attributes := make([]string, 0)
	attributes = append(attributes, "head")
	res2, err := ossClient.Head(guid+"123", attributes)
	if res2 != nil || err != nil {
		t.Log("oss head not exist key fail, res:", res2, "err:", err)
		t.Fail()
	}
}
