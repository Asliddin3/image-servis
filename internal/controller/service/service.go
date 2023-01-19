package service

import (
	"github.com/Asliddin3/image-servis/internal/controller/storage"
	repo "github.com/Asliddin3/image-servis/internal/controller/storage/repo"
	"github.com/Asliddin3/image-servis/pkg/db"
	"github.com/Asliddin3/image-servis/pkg/logger"
)

type ImageService struct {
	Logger     *logger.Logger
	Storage    storage.IStorage
	imageStore repo.DisckImageStore
	// imageStore
}

func NewImageService(l *logger.Logger, stg *db.Postgres, imageStore repo.DisckImageStore) *ImageService {
	return &ImageService{
		Logger:     l,
		Storage:    storage.NewStoragePg(stg),
		imageStore: imageStore,
	}
}
