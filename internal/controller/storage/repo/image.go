package repo

import (
	"context"

	pbi "github.com/Asliddin3/image-servis/genproto/image"
)

type ImageStorageI interface {
	// Createimage(*pbi.CreateimageReq) (*pbi.imageRes, error)
	// Updateimage(*pbi.UpdateimageReq) (*pbi.imageRes, error)
	// CreateStaff(*pbi.CreateStaffReq) (*pbi.CreateStaffRes, error)
	InsertOrUpdateImage(FileName string) error
	GetImages(context.Context) (*pbi.ImagesInfoResponse, error)
	// GetImage(*pbi.StringResponse) (*pbi.ImageInfo, error)
}
