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
	ossClient Client
)

func init() {
	err := godotenv.Overload(".env-oss")
	if err != nil {
		panic(err)
	}

	client, err := New(&Options{
		StorageType:      os.Getenv("StorageType"),
		AccessKeyID:      os.Getenv("AccessKeyID"),
		AccessKeySecret:  os.Getenv("AccessKeySecret"),
		Endpoint:         os.Getenv("Endpoint"),
		Bucket:           os.Getenv("Bucket"),
		Region:           os.Getenv("Region"),
		S3ForcePathStyle: os.Getenv("S3ForcePathStyle") == "true",
		SSL:              os.Getenv("SSL") == "true",
	})

	if err != nil {
		panic(err)
	}

	ossClient = client
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
	if err != nil || head != expectHead {
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
	defer res1.Close()
	byteRes, _ := ioutil.ReadAll(res1)
	if string(byteRes) != content {
		t.Fatal("oss get as reader, readAll error")
	}

	res, err = ossClient.Get(guid, EnableCRCValidation())
	if err != nil || res != content {
		t.Log("oss get content fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestOSS_GetWithMeta(t *testing.T) {
	attributes := make([]string, 0)
	attributes = append(attributes, "head")
	res, meta, err := ossClient.GetWithMeta(guid, attributes)
	if err != nil {
		t.Fatal("oss get content as reader fail, err:", err)
	}
	defer res.Close()
	byteRes, _ := ioutil.ReadAll(res)
	if string(byteRes) != content {
		t.Fatal("oss get as reader, readAll error")
	}

	head, err := strconv.Atoi(meta["head"])
	if err != nil || head != expectHead {
		t.Log("oss get head fail, res:", res, "err:", err)
		t.Fail()
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

func TestOSS_Range(t *testing.T) {
	meta := make(map[string]string)
	err := ossClient.Put(guid, strings.NewReader("123456"), meta)
	if err != nil {
		t.Log("oss put error", err)
		t.Fail()
	}

	res, err := ossClient.Range(guid, 3, 3)
	if err != nil {
		t.Log("oss range error", err)
		t.Fail()
	}

	byteRes, _ := ioutil.ReadAll(res)
	if string(byteRes) != "456" {
		t.Fatalf("oss range as reader, expect:%s, but is %s", "456", string(byteRes))
	}
}

func TestOSS_Exists(t *testing.T) {
	meta := make(map[string]string)
	err := ossClient.Put(guid, strings.NewReader("123456"), meta)
	if err != nil {
		t.Log("oss put error", err)
		t.Fail()
	}

	// test exists
	ok, err := ossClient.Exists(guid)
	if err != nil {
		t.Log("oss Exists error", err)
		t.Fail()
	}
	if !ok {
		t.Log("oss must Exists, but return not exists")
		t.Fail()
	}

	err = ossClient.Del(guid)
	if err != nil {
		t.Log("oss del key fail, err:", err)
		t.Fail()
	}

	// test not exists
	ok, err = ossClient.Exists(guid)
	if err != nil {
		t.Log("oss Exists error", err)
		t.Fail()
	}
	if ok {
		t.Log("oss must not Exists, but return exists")
		t.Fail()
	}
}
