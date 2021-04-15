package awos

/**
Put your environment configuration in ".env-aws"
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
	S3Guid         = "test123"
	S3Content      = "aaaaaa"
	S3ExpectLength = 6
	S3ExpectHead   = 1

	S3CompressGUID    = "test123-snappy"
	S3CompressContent = "snappy-contentsnappy-contentsnappy-contentsnappy-content"
)

var (
	awsClient Client
)

func init() {
	err := godotenv.Overload(".env-aws")
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

	awsClient = client
}

func TestS3_Put(t *testing.T) {
	meta := make(map[string]string)
	meta["head"] = strconv.Itoa(S3ExpectHead)
	meta["length"] = strconv.Itoa(S3ExpectLength)

	err := awsClient.Put(S3Guid, strings.NewReader(S3Content), meta)
	if err != nil {
		t.Log("aws put error", err)
		t.Fail()
	}

	err = awsClient.Put(S3Guid, bytes.NewReader([]byte(S3Content)), meta)
	if err != nil {
		t.Log("aws put error", err)
		t.Fail()
	}
}

func TestS3_CompressAndPut(t *testing.T) {
	meta := make(map[string]string)
	meta["head"] = strconv.Itoa(S3ExpectHead)
	meta["length"] = strconv.Itoa(S3ExpectLength)

	err := awsClient.CompressAndPut(S3CompressGUID, strings.NewReader(S3CompressContent), meta)
	if err != nil {
		t.Log("aws put error", err)
		t.Fail()
	}

	err = awsClient.CompressAndPut(S3CompressGUID, bytes.NewReader([]byte(S3CompressContent)), meta)
	if err != nil {
		t.Log("aws put error", err)
		t.Fail()
	}
}

func TestS3_Head(t *testing.T) {
	attributes := make([]string, 0)
	attributes = append(attributes, "head")
	var res map[string]string
	var err error
	var head int
	var length int

	res, err = awsClient.Head(S3Guid, attributes)
	if err != nil {
		t.Log("aws head error", err)
		t.Fail()
	}

	head, err = strconv.Atoi(res["head"])
	if err != nil || head != S3ExpectHead {
		t.Log("aws get head fail, res:", res, "err:", err)
		t.Fail()
	}

	attributes = append(attributes, "length")
	res, err = awsClient.Head(S3Guid, attributes)
	if err != nil {
		t.Log("aws head error", err)
		t.Fail()
	}

	head, err = strconv.Atoi(res["head"])
	length, err = strconv.Atoi(res["length"])
	if err != nil || head != S3ExpectHead || length != S3ExpectLength {
		t.Log("aws get head fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestS3_Get(t *testing.T) {
	res, err := awsClient.Get(S3Guid)
	if err != nil || res != S3Content {
		t.Log("aws get S3Content fail, res:", res, "err:", err)
		t.Fail()
	}

	res1, err := awsClient.GetAsReader(S3Guid)
	if err != nil {
		t.Fatal("aws get content as reader fail, err:", err)
	}
	defer res1.Close()

	byteRes, _ := ioutil.ReadAll(res1)
	if string(byteRes) != S3Content {
		t.Fatal("aws get as reader, readAll error")
	}
}

func TestS3_GetWithMeta(t *testing.T) {
	attributes := make([]string, 0)
	attributes = append(attributes, "head")
	res, meta, err := awsClient.GetWithMeta(S3Guid, attributes)
	if err != nil {
		t.Fatal("aws get content as reader fail, err:", err)
	}
	defer res.Close()
	byteRes, _ := ioutil.ReadAll(res)
	if string(byteRes) != S3Content {
		t.Fatal("aws get as reader, readAll error")
	}

	head, err := strconv.Atoi(meta["head"])
	if err != nil || head != S3ExpectHead {
		t.Log("aws get head fail, res:", res, "err:", err)
		t.Fail()
	}
}

// compressed content
func TestS3_GetAndDecompress(t *testing.T) {
	res, err := awsClient.GetAndDecompress(S3CompressGUID)
	if err != nil || res != S3CompressContent {
		t.Log("aws get S3 conpressed Content fail, res:", res, "err:", err)
		t.Fail()
	}

	res1, err := awsClient.GetAndDecompressAsReader(S3CompressGUID)
	if err != nil {
		t.Fatal("aws get compressed content as reader fail, err:", err)
	}

	byteRes, error := ioutil.ReadAll(res1)
	if string(byteRes) != S3CompressContent || error != nil {
		t.Fatal("aws get as reader, readAll error0", string(byteRes), error)
	}
}

// non-compressed content
func TestS3_GetAndDecompress2(t *testing.T) {
	res, err := awsClient.GetAndDecompress(S3Guid)
	if err != nil || res != S3Content {
		t.Log("aws get S3Content fail, res:", res, "err:", err)
		t.Fail()
	}

	res1, err := awsClient.GetAndDecompressAsReader(S3Guid)
	if err != nil {
		t.Fatal("aws get content as reader fail, err:", err)
	}

	byteRes, _ := ioutil.ReadAll(res1)
	if string(byteRes) != S3Content {
		t.Fatal("aws get as reader, readAll error")
	}
}

func TestS3_SignURL(t *testing.T) {
	res, err := awsClient.SignURL(S3Guid, 60)
	if err != nil {
		t.Log("oss signUrl fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestS3_ListObject(t *testing.T) {
	res, err := awsClient.ListObject(S3Guid, S3Guid[0:4], "", 10, "")
	if err != nil || len(res) == 0 {
		t.Log("aws list objects fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestS3_Del(t *testing.T) {
	err := awsClient.Del(S3Guid)
	if err != nil {
		t.Log("aws del key fail, err:", err)
		t.Fail()
	}
}

func TestS3_GetNotExist(t *testing.T) {
	res1, err := awsClient.Get(S3Guid + "123")
	if res1 != "" || err != nil {
		t.Log("aws get not exist key fail, res:", res1, "err:", err)
		t.Fail()
	}

	attributes := make([]string, 0)
	attributes = append(attributes, "head")
	res2, err := awsClient.Head(S3Guid+"123", attributes)
	if res2 != nil || err != nil {
		t.Log("aws head not exist key fail, res:", res2, "err:", err, err.Error())
		t.Fail()
	}
}

func TestS3_DelMulti(t *testing.T) {
	keys := []string{"aaa", "bbb", "ccc"}
	for _, key := range keys {
		awsClient.Put(key, strings.NewReader("2333333"), nil)
	}

	err := awsClient.DelMulti(keys)
	if err != nil {
		t.Log("aws del multi keys fail, err:", err)
		t.Fail()
	}

	for _, key := range keys {
		res, err := awsClient.Get(key)
		if res != "" || err != nil {
			t.Logf("key:%s should not be exist", key)
			t.Fail()
		}
	}
}

func TestS3_Range(t *testing.T) {
	meta := make(map[string]string)
	err := awsClient.Put(guid, strings.NewReader("123456"), meta)
	if err != nil {
		t.Log("aws put error", err)
		t.Fail()
	}

	res, err := awsClient.Range(guid, 3, 3)
	if err != nil {
		t.Log("aws range error", err)
		t.Fail()
	}

	byteRes, _ := ioutil.ReadAll(res)
	if string(byteRes) != "456" {
		t.Fatalf("aws range as reader, expect:%s, but is %s", "456", string(byteRes))
	}
}
