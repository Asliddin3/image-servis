package storage

import (
	"github.com/Asliddin3/image-servis/internal/controller/storage/postgres"
	"github.com/Asliddin3/image-servis/internal/controller/storage/repo"
	"github.com/Asliddin3/image-servis/pkg/db"
)

type IStorage interface {
	Image() repo.ImageStorageI
}

type StoragePg struct {
	Db        *db.Postgres
	imageRepo repo.ImageStorageI
}

func NewStoragePg(db *db.Postgres) *StoragePg {
	return &StoragePg{
		Db:        db,
		imageRepo: postgres.NewimageRepo(db),
	}
}

func (s StoragePg) Image() repo.ImageStorageI {
	return s.imageRepo
}
