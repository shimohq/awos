package awos

import (
	"io"
	"os"
	"testing"
)

// TestCompress_gzip
// src_path 原始文件
// target_path 生成的目标文件
// 测试读取文件后压缩 输出到指定path 对比是否正常
func TestCompress_gzip(t *testing.T) {
	path := os.Getenv("src_path")
	source, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	length, err := GetReaderLength(source)
	if err != nil {
		panic(err)
	}
	t.Logf("length %d", length)
	reader, err := DefaultGzipCompressor.Compress(source)
	if err != nil {
		panic(err)
	}
	targetPath := os.Getenv("target_path")
	_, err = os.Stat(targetPath)
	if err == nil {
		os.Remove(targetPath)
	}
	target, err := os.Create(targetPath)
	if err != nil {
		panic(err)
	}
	defer target.Close()
	defer source.Close()
	_, err = io.Copy(target, reader)
	if err != nil {
		panic(err)
	}
}
