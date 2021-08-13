package archive

import (
	"bytes"
	"compress/flate"
	"io"
	"os"

	"github.com/mholt/archiver/v3"
)

// ZipFile implements a method to write files into a zip archive.
// After creating a new file with NewZipFile() you can write to the
// file using z.Write(). You must close the writer when you are done
// and then you can retrieve the file bytes using z.Data()
type ZipFile struct {
	archive archiver.Zip
	buf     *bytes.Buffer
}

// NewZipFile creates a new zip file and opens it ready for writing
func NewZipFile() (*ZipFile, error) {
	z := &ZipFile{
		archive: archiver.Zip{
			FileMethod:             archiver.Store,
			CompressionLevel:       flate.BestSpeed,
			MkdirAll:               true,
			SelectiveCompression:   true,
			ContinueOnError:        false,
			OverwriteExisting:      true,
			ImplicitTopLevelFolder: false,
		},
		buf: bytes.NewBuffer(nil),
	}

	err := z.archive.Create(z.buf)
	if err != nil {
		return nil, err
	}
	return z, nil
}

// CloseWriter closes the internal zip file writer, no files
// should be written to the archive after this is called
func (z *ZipFile) CloseWriter() error {
	return z.archive.Close()
}

// Write writes data to the zip archive
func (z *ZipFile) Write(fileName string, fi os.FileInfo, data io.ReadCloser) error {
	return z.archive.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   fi,
			CustomName: fileName,
		},
		ReadCloser: data,
	})
}

// Data returns the byte content of the zip archive
func (z *ZipFile) Data() []byte {
	return z.buf.Bytes()
}
