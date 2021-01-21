package awos

type putOptions struct {
	contentType     string
	contentEncoding *string
	contentDisposition *string
}

type PutOptions func(options *putOptions)

func PutWithContentType(contentType string) PutOptions {
	return func(options *putOptions) {
		options.contentType = contentType
	}
}

func PutWithContentEncoding(contentEncoding string) PutOptions {
	return func(options *putOptions) {
		options.contentEncoding = &contentEncoding
	}
}

func PutWithContentDisposition(contentDisposition string) PutOptions {
	return func(options *putOptions) {
		options.contentDisposition = &contentDisposition
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
