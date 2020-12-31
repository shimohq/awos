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

func DefaultPutOptions() *putOptions {
	return &putOptions{
		contentType: "text/plain",
	}
}

type getOptions struct {
	contentType     *string
	contentEncoding *string
}

func DefaultGetOptions() *getOptions {
	return &getOptions{}
}

type GetOptions func(options *getOptions)

func GetWithContentType(contentType string) GetOptions {
	return func(options *getOptions) {
		options.contentType = &contentType
	}
}

func GetWithContentEncoding(contentEncoding string) GetOptions {
	return func(options *getOptions) {
		options.contentEncoding = &contentEncoding
	}
}
