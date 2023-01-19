package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Asliddin3/image-servis/genproto/image"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestClientGetImages(t *testing.T) {
	t.Parallel()
	// imageServiceHost := os.Getenv("IMAGE_SERVICE_HOST")
	// imageServicePort := os.Getenv("IMAGE_SERVICE_PORT")
	imageServiceHost := "localhost"
	imageServicePort := "7000"
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", imageServiceHost, imageServicePort), grpc.WithInsecure())
	require.NoError(t, err)
	imageService := image.NewImageServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	_, err = imageService.GetImages(ctx, &image.Empty{})

	require.NoError(t, err)

}

func TestClientUploadImage(t *testing.T) {
	t.Parallel()
	// imageServiceHost := os.Getenv("IMAGE_SERVICE_HOST")
	// imageServicePort := os.Getenv("IMAGE_SERVICE_PORT")
	// fmt.Println("host ",imageServiceHost, imageServicePort)
	imageServiceHost := "localhost"
	imageServicePort := "7000"
	testImageFolder := "../../../tmp"
	imageName := "gohper"
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", imageServiceHost, imageServicePort), grpc.WithInsecure())
	require.NoError(t, err)
	imageService := image.NewImageServiceClient(conn)
	imagePath := fmt.Sprintf("%s/%s.jpg", testImageFolder, imageName)
	file, err := os.Open(imagePath)
	require.NoError(t, err)
	defer file.Close()
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	// defer cancel()
	stream, err := imageService.UploadFile(context.Background())
	require.NoError(t, err)

	imageType := filepath.Ext(imagePath)
	req := &image.UploadImageRequest{
		Request: &image.UploadImageRequest_Info{
			Info: &image.ImageInfo{
				ImageName: imageName,
				ImageData: imageType,
			},
		},
	}
	err = stream.Send(req)
	require.NoError(t, err)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			fmt.Println("EOF in for")
			break
		}
		require.NoError(t, err)
		size += n

		req := &image.UploadImageRequest{
			Request: &image.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		fmt.Println("reader err", err, "stream error ", err)
		require.NoError(t, err)
	}
	_, err = stream.CloseAndRecv()
	require.NoError(t, err)
	// time.Sleep(time.Duration(time.Second * 2))
	savedImagePath := fmt.Sprintf("%s/%s%s", "../../../img", imageName, imageType)

	require.FileExists(t, savedImagePath)
	// require.NoError(t, os.Remove(savedImagePath))
}

func TestClientDownloadImage(t *testing.T) {
	t.Parallel()

	// imageServiceHost := os.Getenv("IMAGE_SERVICE_HOST")
	// imageServicePort := os.Getenv("IMAGE_SERVICE_PORT")
	imageServiceHost := "localhost"
	imageServicePort := "7000"
	testImageFolder := "../../../tmp"
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", imageServiceHost, imageServicePort), grpc.WithInsecure())
	require.NoError(t, err)
	imageName := "laptop"
	imageType := "jpg"
	imageService := image.NewImageServiceClient(conn)
	stream, err := imageService.DownloadFile(context.Background(), &image.ImageInfo{
		ImageName: imageName,
		ImageData: imageType})
	require.NoError(t, err)
	imageData := bytes.Buffer{}
	imageSize := 0
	for {
		err := contextError(stream.Context())
		if err != nil {
			break
		}

		req, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("no more data")
			break
		}
		require.NoError(t, err)
		if err != nil {
			fmt.Printf("cannot receive chunk data %s\n", err.Error())
			break
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		imageSize += size

		_, err = imageData.Write(chunk)
		require.NoError(t, err)
		if err != nil {
			break
		}
	}

	imagePath := fmt.Sprintf("%s/%s.%s", testImageFolder, imageName, imageType)
	fmt.Println(imagePath)
	file, err := os.Create(imagePath)
	defer file.Close()
	require.NoError(t, err)

	_, err = imageData.WriteTo(file)

	require.NoError(t, err)
	savedImagePath := fmt.Sprintf("%s/%s.%s", testImageFolder, "laptop", imageType)
	require.FileExists(t, savedImagePath)
	// require.NoError(t, os.Remove(savedImagePath))
}
