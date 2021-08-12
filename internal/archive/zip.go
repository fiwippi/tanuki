package archive

import (
	"bytes"
	"compress/flate"
	"io"
	"os"

	"github.com/mholt/archiver/v3"
)

type ZipFile struct {
	archive archiver.Zip
	buf     *bytes.Buffer
}

func NewZipFile() (*ZipFile, error) {
	z := &ZipFile{
		archive: archiver.Zip{
			CompressionLevel:       flate.BestSpeed,
			MkdirAll:               true,
			SelectiveCompression:   true,
			ContinueOnError:        false,
			OverwriteExisting:      true,
			ImplicitTopLevelFolder: false,
			FileMethod:             archiver.Store,
		},
		buf: bytes.NewBuffer(nil),
	}

	err := z.archive.Create(z.buf)
	if err != nil {
		return nil, err
	}
	return z, nil
}

func (z *ZipFile) CloseWriter() error {
	return z.archive.Close()
}

func (z *ZipFile) Write(fileName string, fi os.FileInfo, data io.ReadCloser) error {
	return z.archive.Write(archiver.File{
		FileInfo: archiver.FileInfo{
			FileInfo:   fi,
			CustomName: fileName,
		},
		ReadCloser: data,
	})
}

func (z *ZipFile) Data() []byte {
	return z.buf.Bytes()
}
