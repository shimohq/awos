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
	return &gzipReadSeeker{
		reader: reader,
	}, nil
}

func (g *GzipCompressor) ContentEncoding() string {
	return compressTypeGzip
}

type gzipReadSeeker struct {
	reader io.ReadSeeker
}

func (crs *gzipReadSeeker) Read(p []byte) (n int, err error) {
	// 读取原始数据
	n, err = crs.reader.Read(p)
	if err != nil && err != io.EOF {
		return n, err
	}
	if n == 0 {
		return 0, err
	}
	var compressedBuffer bytes.Buffer
	gw := gzip.NewWriter(&compressedBuffer)
	// 压缩读取的数据
	_, err = gw.Write(p[:n])
	if err != nil {
		_ = gw.Close()
		return n, err
	}
	if err = gw.Close(); err != nil {
		return 0, err
	}
	// 将压缩后的数据返回给调用者
	n = copy(p, compressedBuffer.Bytes())
	compressedBuffer.Reset()
	return n, err
}

func (crs *gzipReadSeeker) Seek(offset int64, whence int) (int64, error) {
	// 调用原始ReadSeeker的Seek方法
	return crs.reader.Seek(offset, whence)
}

var DefaultGzipCompressor = &GzipCompressor{}

func GetReaderLength(reader io.ReadSeeker) (int64, error) {
	// 保存当前的读写位置
	originalPos, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	// 移动到文件末尾以获取字节长度
	length, err := reader.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	// 恢复原始读写位置
	_, err = reader.Seek(originalPos, io.SeekStart)
	if err != nil {
		return 0, err
	}

	return length, nil
}
