package repo

import (
	"bytes"

	pbc "github.com/Asliddin3/image-servis/genproto/image"
)

type DisckImageStore interface {
	Save(ImageName string, ImageType string, ImageData bytes.Buffer) error
	GetImage(ImageName, ImageType string, stream pbc.ImageService_DownloadFileServer) error
}
