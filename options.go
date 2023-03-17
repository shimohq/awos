package awos

import "time"

type putOptions struct {
	contentType        string
	contentEncoding    *string
	contentDisposition *string
	cacheControl       *string
	expires            *time.Time
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

func PutWithCacheControl(cacheControl string) PutOptions {
	return func(options *putOptions) {
		options.cacheControl = &cacheControl
	}
}

func PutWithExpireTime(expires time.Time) PutOptions {
	return func(options *putOptions) {
		options.expires = &expires
	}
}

func DefaultPutOptions() *putOptions {
	return &putOptions{
		contentType: "text/plain",
	}
}

type getOptions struct {
	contentType         *string
	contentEncoding     *string
	enableCRCValidation bool
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

func EnableCRCValidation() GetOptions {
	return func(options *getOptions) {
		options.enableCRCValidation = true
	}
}

type SignOptions func(options *signOptions)

func SignWithProcess(process string) SignOptions {
	return func(options *signOptions) {
		options.process = &process
	}
}

type signOptions struct {
	process *string
}

func DefaultSignOptions() *signOptions {
	return &signOptions{}
}
