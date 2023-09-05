package awos

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/service/s3"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

type HeadGetObjectOutputWrapper struct {
	getObjectOutput  *s3.GetObjectOutput
	headObjectOutput *s3.HeadObjectOutput
}

func (h *HeadGetObjectOutputWrapper) getContentType() *string {
	if h.getObjectOutput != nil {
		return h.getObjectOutput.ContentType
	}
	return h.headObjectOutput.ContentType
}

func (h *HeadGetObjectOutputWrapper) getContentEncoding() *string {
	if h.getObjectOutput != nil {
		return h.getObjectOutput.ContentEncoding
	}
	return h.headObjectOutput.ContentEncoding
}

func (h *HeadGetObjectOutputWrapper) getContentLength() *string {
	if h.getObjectOutput != nil {
		if h.getObjectOutput.ContentLength != nil {
			clStr := strconv.FormatInt(*h.getObjectOutput.ContentLength, 10)
			return &clStr
		} else {
			// 异常情况，理论上不应该出现
			logger.Warn("getObjectOutput.ContentLength is nil", "getObjectOutput", h.getObjectOutput)
		}
	}
	clStr := strconv.FormatInt(*h.headObjectOutput.ContentLength, 10)
	return &clStr
}

func (h *HeadGetObjectOutputWrapper) getContentDisposition() *string {
	if h.getObjectOutput != nil {
		return h.getObjectOutput.ContentDisposition
	}
	return h.headObjectOutput.ContentDisposition
}

func (h *HeadGetObjectOutputWrapper) metaData() map[string]*string {
	if h.getObjectOutput != nil {
		return h.getObjectOutput.Metadata
	}
	return h.headObjectOutput.Metadata
}

func mergeHttpStandardHeaders(output *HeadGetObjectOutputWrapper) map[string]*string {
	res := make(map[string]*string)
	for k, v := range output.metaData() {
		res[k] = v
	}

	res["Content-Length"] = output.getContentLength()
	res["Content-Encoding"] = output.getContentEncoding()
	res["Content-Type"] = output.getContentType()
	res["Content-Disposition"] = output.getContentDisposition()

	return res
}
