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
	AWSGuid         = "test123"
	AWSContent      = "aaaaaa"
	AWSExpectLength = 6
	AWSExpectHead   = 1

	AWSCompressGUID    = "test123-snappy"
	AWSCompressContent = "snappy-contentsnappy-contentsnappy-contentsnappy-content"
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

func TestAWS_Put(t *testing.T) {
	meta := make(map[string]string)
	meta["head"] = strconv.Itoa(AWSExpectHead)
	meta["length"] = strconv.Itoa(AWSExpectLength)

	err := awsClient.Put(AWSGuid, strings.NewReader(AWSContent), meta)
	if err != nil {
		t.Log("aws put error", err)
		t.Fail()
	}

	err = awsClient.Put(AWSGuid, bytes.NewReader([]byte(AWSContent)), meta)
	if err != nil {
		t.Log("aws put error", err)
		t.Fail()
	}
}

func TestAWS_CompressAndPut(t *testing.T) {
	meta := make(map[string]string)
	meta["head"] = strconv.Itoa(AWSExpectHead)
	meta["length"] = strconv.Itoa(AWSExpectLength)

	err := awsClient.CompressAndPut(AWSCompressGUID, strings.NewReader(AWSCompressContent), meta)
	if err != nil {
		t.Log("aws put error", err)
		t.Fail()
	}

	err = awsClient.CompressAndPut(AWSCompressGUID, bytes.NewReader([]byte(AWSCompressContent)), meta)
	if err != nil {
		t.Log("aws put error", err)
		t.Fail()
	}
}

func TestAWS_Head(t *testing.T) {
	attributes := make([]string, 0)
	attributes = append(attributes, "head")
	var res map[string]string
	var err error
	var head int
	var length int

	res, err = awsClient.Head(AWSGuid, attributes)
	if err != nil {
		t.Log("aws head error", err)
		t.Fail()
	}

	head, err = strconv.Atoi(res["head"])
	if err != nil || head != 1 {
		t.Log("aws get head fail, res:", res, "err:", err)
		t.Fail()
	}

	attributes = append(attributes, "length")
	res, err = awsClient.Head(AWSGuid, attributes)
	if err != nil {
		t.Log("aws head error", err)
		t.Fail()
	}

	head, err = strconv.Atoi(res["head"])
	length, err = strconv.Atoi(res["length"])
	if err != nil || head != AWSExpectHead || length != AWSExpectLength {
		t.Log("aws get head fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestAWS_Get(t *testing.T) {
	res, err := awsClient.Get(AWSGuid)
	if err != nil || res != AWSContent {
		t.Log("aws get AWSContent fail, res:", res, "err:", err)
		t.Fail()
	}

	res1, err := awsClient.GetAsReader(AWSGuid)
	if err != nil {
		t.Fatal("aws get content as reader fail, err:", err)
	}

	byteRes, _ := ioutil.ReadAll(res1)
	if string(byteRes) != AWSContent {
		t.Fatal("aws get as reader, readAll error")
	}
}

// compressed content
func TestAWS_GetAndDecompress(t *testing.T) {
	res, err := awsClient.GetAndDecompress(AWSCompressGUID)
	if err != nil || res != AWSCompressContent {
		t.Log("aws get AWS conpressed Content fail, res:", res, "err:", err)
		t.Fail()
	}

	res1, err := awsClient.GetAndDecompressAsReader(AWSCompressGUID)
	if err != nil {
		t.Fatal("aws get compressed content as reader fail, err:", err)
	}

	byteRes, error := ioutil.ReadAll(res1)
	if string(byteRes) != AWSCompressContent || error != nil {
		t.Fatal("aws get as reader, readAll error0", string(byteRes), error)
	}
}

// non-compressed content
func TestAWS_GetAndDecompress2(t *testing.T) {
	res, err := awsClient.GetAndDecompress(AWSGuid)
	if err != nil || res != AWSContent {
		t.Log("aws get AWSContent fail, res:", res, "err:", err)
		t.Fail()
	}

	res1, err := awsClient.GetAndDecompressAsReader(AWSGuid)
	if err != nil {
		t.Fatal("aws get content as reader fail, err:", err)
	}

	byteRes, _ := ioutil.ReadAll(res1)
	if string(byteRes) != AWSContent {
		t.Fatal("aws get as reader, readAll error")
	}
}

func TestAWS_SignURL(t *testing.T) {
	res, err := awsClient.SignURL(AWSGuid, 60)
	if err != nil {
		t.Log("oss signUrl fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestAWS_ListObject(t *testing.T) {
	res, err := awsClient.ListObject(AWSGuid, AWSGuid[0:4], "", 10, "")
	if err != nil || len(res) == 0 {
		t.Log("aws list objects fail, res:", res, "err:", err)
		t.Fail()
	}
}

func TestAWS_Del(t *testing.T) {
	err := awsClient.Del(AWSGuid)
	if err != nil {
		t.Log("aws del key fail, err:", err)
		t.Fail()
	}
}

func TestAWS_GetNotExist(t *testing.T) {
	res1, err := awsClient.Get(AWSGuid + "123")
	if res1 != "" || err != nil {
		t.Log("aws get not exist key fail, res:", res1, "err:", err)
		t.Fail()
	}

	attributes := make([]string, 0)
	attributes = append(attributes, "head")
	res2, err := awsClient.Head(AWSGuid+"123", attributes)
	if res2 != nil || err != nil {
		t.Log("aws head not exist key fail, res:", res2, "err:", err, err.Error())
		t.Fail()
	}
}

func TestAWS_DelMulti(t *testing.T) {
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

func TestAWS_Range(t *testing.T) {
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
