package awos

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)

func TestHeadGetObjectOutputWrapper_getContentLength(t *testing.T) {
	type fields struct {
		getObjectOutput  *s3.GetObjectOutput
		headObjectOutput *s3.HeadObjectOutput
	}
	tests := []struct {
		name   string
		fields fields
		want   *string
	}{
		// TODO: Add test cases.
		{
			name: "getObjectOutput.ContentLength is nil",
			fields: fields{
				getObjectOutput: &s3.GetObjectOutput{
					ContentLength: nil,
				},
				headObjectOutput: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HeadGetObjectOutputWrapper{
				getObjectOutput:  tt.fields.getObjectOutput,
				headObjectOutput: tt.fields.headObjectOutput,
			}
			assert.Equalf(t, tt.want, h.getContentLength(), "getContentLength()")
		})
	}
}
