package awos

/**
Put your environment configuration in ".env-oss"
*/

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/joho/godotenv"
)

const (
	guid         = "test123"
	content      = "aaaaaa"
	expectLength = 6
	expectHead   = 1

	compressGUID    = "test123-snappy"
	compressContent = "snappy-contentsnappy-contentsnappy-contentsnappy-content"
)

var (
	ossClient *OSS
)

func init() {
	err := godotenv.Overload(".env-oss")
	if err != nil {
		panic(err)
	}

	client, err := oss.New(os.Getenv("Endpoint"), os.Getenv("AccessKeyId"),
		os.Getenv("AccessKeySecret"))
	if err != nil {
		panic(err)
	}

	bucket, err := client.Bucket(os.Getenv("Bucket"))
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

	err := ossClient.Put(guid, strings.NewReader(content), meta)
	if err != nil {
		t.Log("oss put error", err)
		t.Fail()
	}

	err = ossClient.Put(guid, bytes.NewReader([]byte(content)), meta)
	if err != nil {
		t.Log("oss put byte array error", err)
		t.Fail()
	}
}

func TestOSS_CompressAndPut(t *testing.T) {
	meta := make(map[string]string)
	meta["head"] = strconv.Itoa(expectHead)
	meta["length"] = strconv.Itoa(expectLength)

	err := ossClient.CompressAndPut(compressGUID, strings.NewReader(compressContent), meta)
	if err != nil {
		t.Log("oss put error", err)
		t.Fail()
	}

	err = ossClient.CompressAndPut(compressGUID, bytes.NewReader([]byte(compressContent)), meta)
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

	res1, err := ossClient.GetAsReader(guid)
	if err != nil {
		t.Fatal("oss get content as reader fail, err:", err)
	}

	byteRes, _ := ioutil.ReadAll(res1)
	if string(byteRes) != content {
		t.Fatal("oss get as reader, readAll error")
	}
}

func TestOSS_GetAndDecompress(t *testing.T) {
	res, err := ossClient.GetAndDecompress(compressGUID)
	if err != nil || res != compressContent {
		t.Log("aws get oss content fail, res:", res, "err:", err)
		t.Fail()
	}

	res1, err := ossClient.GetAndDecompressAsReader(compressGUID)
	if err != nil {
		t.Fatal("oss get content as reader fail, err:", err)
	}

	byteRes, _ := ioutil.ReadAll(res1)
	if string(byteRes) != compressContent {
		t.Fatal("oss get as reader, readAll error")
	}
}

func TestOSS_GetAndDecompress2(t *testing.T) {
	res, err := ossClient.GetAndDecompress(guid)
	if err != nil || res != content {
		t.Log("aws get oss content fail, res:", res, "err:", err)
		t.Fail()
	}

	res1, err := ossClient.GetAndDecompressAsReader(guid)
	if err != nil {
		t.Fatal("oss get content as reader fail, err:", err)
	}

	byteRes, _ := ioutil.ReadAll(res1)
	if string(byteRes) != content {
		t.Fatal("oss get as reader, readAll error")
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
