package archive

import (
	"archive/zip"
	"bytes"
	"io"
)

type ZipFile struct {
	buf    *bytes.Buffer
	writer *zip.Writer
}

func (z *ZipFile) Write(name string, r io.ReadCloser) error {
	if z.buf == nil {
		z.buf = new(bytes.Buffer)
		z.writer = zip.NewWriter(z.buf)
	}

	// Create file in archive
	f, err := z.writer.Create(name)
	if err != nil {
		return err
	}

	// Write data to the file
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (z *ZipFile) Close() error {
	return z.writer.Close()
}

func (z *ZipFile) Bytes() []byte {
	return z.buf.Bytes()
}
