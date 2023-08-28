package awos

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"
)

var (
	compressTypeGzip = "gzip"
	compressorsMu    sync.RWMutex
	compressors      = make(map[string]Compressor)
)

func Register(comp Compressor) {
	compressorsMu.Lock()
	defer compressorsMu.Unlock()
	compressors[comp.ContentEncoding()] = comp
}

type Compressor interface {
	Compress(reader io.ReadSeeker) (gzipReader io.ReadSeeker, err error)
	ContentEncoding() string
}

type GzipCompressor struct {
}

func (g *GzipCompressor) Compress(reader io.ReadSeeker) (gzipReader io.ReadSeeker, err error) {
	//	TODO buffer limit
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	_, err = io.Copy(gzipWriter, reader)
	if err != nil {
		return nil, err
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buffer.Bytes()), nil
}

func (g *GzipCompressor) ContentEncoding() string {
	return compressTypeGzip
}

var DefaultGzipCompressor = &GzipCompressor{}

func GetReaderLength(reader io.ReadSeeker) (io.ReadSeeker, int, error) {
	all, err := io.ReadAll(reader)
	if err != nil {
		return nil, 0, err
	}
	return bytes.NewReader(all), len(all), nil
}
