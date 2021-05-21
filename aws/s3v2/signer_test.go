package s3v2

import (
	"net/http"
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/awstesting"
)

// Tests for 'func TestResourceListSorting(t *testing.T)'.
func TestResourceListSorting(t *testing.T) {
	sortedResourceList := make([]string, len(resourceList))
	copy(sortedResourceList, resourceList)
	sort.Strings(sortedResourceList)
	for i := 0; i < len(resourceList); i++ {
		if resourceList[i] != sortedResourceList[i] {
			t.Errorf("Expected resourceList[%d] = \"%s\", resourceList is not correctly sorted.", i, sortedResourceList[i])
			break
		}
	}
}

func TestPreSignRequest(t *testing.T) {
	var testCases = []struct {
		endpointURL       string
		accessKeyID       string
		secretAccessKey   string
		sessionToken      string
		signTime          time.Time
		expires           time.Duration
		path              string
		expectedExpires   string
		expectedSignature string
	}{
		{
			endpointURL:       "https://s3v2-compatible.com/",
			accessKeyID:       "AKID",
			secretAccessKey:   "SECRET",
			sessionToken:      "SESSION",
			signTime:          time.Unix(1621836642, 0),
			expires:           1 * time.Minute,
			path:              "/awostest/test123",
			expectedExpires:   "1621836702",
			expectedSignature: "aIShchTjP5Ly7bH0NtX+3yXtPjU=",
		},
	}

	for _, testCase := range testCases {
		creds := credentials.NewStaticCredentials(
			testCase.accessKeyID,
			testCase.secretAccessKey,
			testCase.sessionToken,
		)

		svc := awstesting.NewClient(&aws.Config{
			Credentials: creds,
			Region:      aws.String("cn-north-1"),
		})

		req := svc.NewRequest(
			&request.Operation{
				Name:       "OpName",
				HTTPMethod: http.MethodGet,
				HTTPPath:   testCase.path,
			},
			nil,
			nil,
		)
		req.Time = testCase.signTime
		req.ExpireTime = testCase.expires

		PreSignRequest(req)
		if req.Error != nil {
			t.Fatalf("expect no error, got %v", req.Error)
		}

		reqQuery := req.HTTPRequest.URL.Query()

		if e, a := 3, len(reqQuery); e != a {
			t.Errorf("expect %v, got %v", e, a)
		}
		if e, a := testCase.accessKeyID, reqQuery.Get("AWSAccessKeyId"); e != a {
			t.Errorf("expect %v, got %v", e, a)
		}
		if e, a := testCase.expectedExpires, reqQuery.Get("Expires"); e != a {
			t.Errorf("expect %v, got %v", e, a)
		}
		if e, a := testCase.expectedSignature, reqQuery.Get("Signature"); e != a {
			t.Errorf("expect %v, got %v", e, a)
		}
	}

}

// Tests validate the URL path encoder.
func TestEncodePath(t *testing.T) {
	testCases := []struct {
		// Input.
		inputStr string
		// Expected result.
		result string
	}{
		{"thisisthe%url", "thisisthe%25url"},
		{"本語", "%E6%9C%AC%E8%AA%9E"},
		{"本語.1", "%E6%9C%AC%E8%AA%9E.1"},
		{">123", "%3E123"},
		{"myurl#link", "myurl%23link"},
		{"space in url", "space%20in%20url"},
		{"url+path", "url%2Bpath"},
		{"url/path", "url/path"},
	}

	for i, testCase := range testCases {
		result := encodePath(testCase.inputStr)
		if testCase.result != result {
			t.Errorf("Test %d: Expected queryEncode result to be \"%s\", but found it to be \"%s\" instead", i+1, testCase.result, result)
		}
	}
}
