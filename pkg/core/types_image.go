package core

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

type ImageType int

const (
	ImagePNG ImageType = iota
	ImageJPEG
	ImageGIF
	ImageWEBP
	ImageTIFF
	ImageBMP
)

func (it ImageType) MimeType() string {
	return [...]string{"image/png", "image/jpeg", "image/gif", "image/webp", "image/tiff", "image/bmp"}[it]
}

func (it ImageType) String() string {
	return [...]string{"png", "jpg", "gif", "webp", "tiff", "bmp"}[it]
}

func (it ImageType) Decode(r io.Reader) (image.Image, error){
	switch it {
	case ImagePNG:
		return png.Decode(r)
	case ImageJPEG:
		 return jpeg.Decode(r)
	case ImageGIF:
		return gif.Decode(r)
	case ImageWEBP:
 		return webp.Decode(r)
 	case ImageTIFF:
 		return tiff.Decode(r)
 	case ImageBMP:
 		return bmp.Decode(r)
 	}

	panic(fmt.Sprintf("invalid image type: '%d'", it))
}

func EncodeJPEG(img image.Image) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := jpeg.Encode(buf, img, nil)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func GetImageType(ext string) (ImageType, error) {
	ext = strings.TrimPrefix(ext, ".")
	ext = strings.ToLower(ext)

	switch ext {
	case "png":
		return ImagePNG, nil
	case "jpeg", "jpg":
		return ImageJPEG, nil
	case "gif":
		return ImageGIF, nil
	case "webp":
		return ImageWEBP, nil
	case "tiff":
		return ImageTIFF, nil
	case "bmp":
		return ImageBMP, nil
	}

	return -1, fmt.Errorf("invalid image type: '%s'", ext)
}
