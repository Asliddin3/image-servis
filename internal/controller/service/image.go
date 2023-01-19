package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	pbi "github.com/Asliddin3/image-servis/genproto/image"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (b *ImageService) GetImages(ctx context.Context, _ *pbi.Empty) (*pbi.ImagesInfoResponse, error) {
	res, err := b.Storage.Image().GetImages(ctx)
	if err != nil {
		return nil, logError(status.Errorf(codes.Unknown, "cannot get images from postgres %v", err))
	}
	return res, nil
}

func (b *ImageService) UploadFile(stream pbi.ImageService_UploadFileServer) error {
	req, err := stream.Recv()
	fmt.Println(req.GetInfo(), err)
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info"))
	}
	imageName := req.GetInfo().GetImageName()
	imageType := req.GetInfo().GetImageData()
	fmt.Printf("image-name %s %s", imageName, imageType)

	imageData := bytes.Buffer{}

	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}
		log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			fmt.Println(logError(status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err)))
			break
		}

		chunk := req.GetChunkData()

		// log.Printf("received a chunk with size: %d", size)

		// write slowly
		// time.Sleep(time.Second)

		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))
		}
	}
	// chDone := make(chan struct{})
	err = b.imageStore.Save(imageName, imageType, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image in folder: %v", err))
	}

	err = b.Storage.Image().InsertOrUpdateImage(imageName, imageType)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot insert image to postgres %v", err))
	}
	err = stream.SendAndClose(&pbi.StringResponse{Message: "finished sucesfully"})
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot send msg to client %v", err))
	}
	return nil
}

func (b *ImageService) DownloadFile(req *pbi.ImageInfo, stream pbi.ImageService_DownloadFileServer) error {
	err := b.imageStore.GetImage(req.ImageName, req.ImageData, stream)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot get image from folder: %v", err))
	}
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return logError(status.Error(codes.Canceled, "request is canceled"))
	case context.DeadlineExceeded:
		return logError(status.Error(codes.DeadlineExceeded, "deadline is exceeded"))
	default:
		return nil
	}
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}
