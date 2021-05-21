package utils

import "testing"

type addSchemeTestCase struct {
	endpoint string
	ssl      bool
	expected string
}

func TestAddScheme(t *testing.T) {
	testCases := []addSchemeTestCase{
		{
			endpoint: "endpoint1.com",
			ssl:      true,
			expected: "https://endpoint1.com",
		},
		{
			endpoint: "endpoint1.com",
			ssl:      false,
			expected: "http://endpoint1.com",
		},
	}

	for _, testCase := range testCases {
		added := addScheme(testCase.endpoint, testCase.ssl)
		if added != testCase.expected {
			t.Errorf("should be: %s", testCase.expected)
		}
	}
}

type ossHostTestCase struct {
	before string
	ssl    bool
	after  string
}

func TestToPublicOSSHost(t *testing.T) {
	testCases := []ossHostTestCase{
		{
			before: "https://oss-cn-hangzhou-internal.aliyuncs.com",
			ssl:    false,
			after:  "http://oss-cn-hangzhou.aliyuncs.com",
		},
		{
			before: "http://oss-cn-hangzhou-internal.aliyuncs.com",
			ssl:    true,
			after:  "https://oss-cn-hangzhou.aliyuncs.com",
		},
		{
			before: "http://oss-cn-hangzhou.aliyuncs.com",
			ssl:    true,
			after:  "https://oss-cn-hangzhou.aliyuncs.com",
		},
		{
			before: "https://oss-cn-hangzhou.aliyuncs.com",
			ssl:    false,
			after:  "http://oss-cn-hangzhou.aliyuncs.com",
		},
	}

	for _, testCase := range testCases {
		replacer := ToPublicOSSHost(testCase.ssl)
		if replacer(testCase.before) != testCase.after {
			t.Errorf("should be: %s\n", testCase.after)
		}
	}
}
