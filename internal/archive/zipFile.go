package archive

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
)

// ZipFile implements a method to write files into a zip archive.
// After creating a new file with NewZipFile() you can write to the
// file using z.Write(). You must close the writer when you are done
// and then you can retrieve the file bytes using z.Data()
type ZipFile struct {
	buf    *bytes.Buffer
	writer *zip.Writer
	closed bool
}

// NewZipFile creates a new zip file and opens it ready for writing
func NewZipFile() (*ZipFile, error) {
	b := new(bytes.Buffer)
	z := &ZipFile{
		buf:    b,
		writer: zip.NewWriter(b),
	}

	return z, nil
}

func (z *ZipFile) Write(name string, r io.ReadCloser) error {
	f, err := z.writer.Create(name)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(r)
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
	z.closed = true
	return z.writer.Close()
}

func (z *ZipFile) Data() []byte {
	if !z.closed {
		z.writer.Close()
	}
	return z.buf.Bytes()
}
