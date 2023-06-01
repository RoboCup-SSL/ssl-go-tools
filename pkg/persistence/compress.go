package persistence

import (
	"compress/gzip"
	"io"
	"os"
)

func Compress(inFile, outFile string) error {
	reader, err := os.OpenFile(inFile, os.O_RDONLY, 0660)
	writer, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	gzipWriter := gzip.NewWriter(writer)
	if _, err := io.Copy(gzipWriter, reader); err != nil {
		return err
	}
	if err := reader.Close(); err != nil {
		return err
	}
	if err := gzipWriter.Close(); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return nil
}
