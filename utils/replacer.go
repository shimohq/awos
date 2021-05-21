package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/shimohq/awos/v3"
)

var schemeRegExp = regexp.MustCompile("^([^:]+)://")

func addScheme(endpoint string, ssl bool) string {
	if schemeRegExp.MatchString(endpoint) {
		return endpoint
	}

	var scheme string
	if ssl {
		scheme = "https"
	} else {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s", scheme, endpoint)
}

func ReplaceEndpoint(toEndpoint string, ssl bool) awos.PresignEndpointReplacer {
	finalEndpoint := addScheme(toEndpoint, ssl)
	return func(endpointBefore string) string {
		return finalEndpoint
	}
}

func ToPublicOSSHost(ssl bool) awos.PresignedURLReplacer {
	return func(presigned string) string {
		u, _ := url.Parse(presigned)

		u.Host = strings.Replace(u.Host, "-internal.", ".", 1)
		if ssl {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}

		return u.String()
	}
}
