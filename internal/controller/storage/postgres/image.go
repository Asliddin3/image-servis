package postgres

import (
	"context"

	pbi "github.com/Asliddin3/image-servis/genproto/image"
)

func (b *imageRepo) GetImages(ctx context.Context) (*pbi.ImagesInfoResponse, error) {
	res, _, err := b.Db.Builder.Select("name", "cast(created_at AS TEXT)", "cast(updated_at AS TEXT)").From("images").ToSql()
	if err != nil {
		return &pbi.ImagesInfoResponse{}, err
	}
	rows, err := b.Db.Pool.Query(ctx, res)
	if err != nil {
		return &pbi.ImagesInfoResponse{}, err
	}
	imagesResp := &pbi.ImagesInfoResponse{}
	for rows.Next() {
		imageResp := &pbi.ImageInfoResponse{}
		err = rows.Scan(&imageResp.ImageName, &imageResp.CreatedAt, &imageResp.UpdatedAt)
		if err != nil {
			return &pbi.ImagesInfoResponse{}, err
		}
		imagesResp.Images = append(imagesResp.Images, imageResp)
	}
	return imagesResp, nil
}

func (b *imageRepo) InsertOrUpdateImage(fileName string) error {
	res := `
	INSERT INTO images VALUES ($1)
	ON CONFLICT (name) DO UPDATE SET updated_at = current_timestamp;
	`
	_, err := b.Db.Pool.Exec(context.Background(), res, fileName)

	if err != nil {
		return err
	}
	return nil
}
