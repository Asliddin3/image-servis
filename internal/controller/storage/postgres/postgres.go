package postgres

import "github.com/Asliddin3/image-servis/pkg/db"

type imageRepo struct {
	Db *db.Postgres
}

func NewimageRepo(db *db.Postgres) *imageRepo {
	return &imageRepo{Db: db}
}
