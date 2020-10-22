package awos

import "io"

// CombinedReadCloser combined a ReadCloser and a Readers to a new ReaderCloser
// which will read from reader and close origin closer
type CombinedReadCloser struct {
	ReadCloser io.ReadCloser
	Reader     io.Reader
}

func (combined CombinedReadCloser) Read(b []byte) (int, error) {
	return combined.Reader.Read(b)
}

// Close origin ReaderCloser
func (combined CombinedReadCloser) Close() error {
	return combined.ReadCloser.Close()
}
