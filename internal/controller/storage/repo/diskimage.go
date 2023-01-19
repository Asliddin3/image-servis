package repo

import (
	"bytes"

	pbc "github.com/Asliddin3/image-servis/genproto/image"
)

type DisckImageStore interface {
	Save(FileName string, ImageData bytes.Buffer) error
	GetImage(FileName string, stream pbc.ImageService_DownloadFileServer) error
}
