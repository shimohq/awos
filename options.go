package awos

type putOptions struct {
	contentType string
}

type PutOptions func(options *putOptions)

func PutWithContentType(contentType string) PutOptions {
	return func(options *putOptions) {
		options.contentType = contentType
	}
}

var DefaultPutOptions = &putOptions{
	contentType: "text/plain",
}
