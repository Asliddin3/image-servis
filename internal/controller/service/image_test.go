package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
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
	imageServiceHost := "localhost"
	imageServicePort := "7000"
	testImageFolder := "../../../tmp"
	fileName := "gopher1.png"
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", imageServiceHost, imageServicePort), grpc.WithInsecure())
	require.NoError(t, err)
	imageService := image.NewImageServiceClient(conn)
	imagePath := fmt.Sprintf("%s/%s", testImageFolder, fileName)
	file, err := os.Open(imagePath)
	require.NoError(t, err)
	defer file.Close()
	stream, err := imageService.UploadFile(context.Background())
	require.NoError(t, err)

	req := &image.UploadImageRequest{
		Request: &image.UploadImageRequest_Info{
			Info: &image.ImageInfo{
				FileName: fileName,
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
	savedImagePath := fmt.Sprintf("%s/%s", "../../../img", fileName)

	require.FileExists(t, savedImagePath)
}

func TestClientDownloadImage(t *testing.T) {
	t.Parallel()
	imageServiceHost := "localhost"
	imageServicePort := "7000"
	testImageFolder := "../../../tmp"
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", imageServiceHost, imageServicePort), grpc.WithInsecure())
	require.NoError(t, err)
	fileName := "laptop0.jpg"
	imageService := image.NewImageServiceClient(conn)
	stream, err := imageService.DownloadFile(context.Background(), &image.ImageInfo{
		FileName: fileName})
	require.NoError(t, err)
	imageData := bytes.Buffer{}
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

		_, err = imageData.Write(chunk)
		require.NoError(t, err)
		if err != nil {
			break
		}
	}
	err = stream.CloseSend()
	require.NoError(t, err)
	imagePath := fmt.Sprintf("%s/%s", testImageFolder, fileName)
	fmt.Println(imagePath)
	file, err := os.Create(imagePath)
	defer file.Close()
	require.NoError(t, err)

	_, err = imageData.WriteTo(file)

	require.NoError(t, err)
	savedImagePath := fmt.Sprintf("%s/%s", testImageFolder, fileName)
	require.FileExists(t, savedImagePath)
	// require.NoError(t, os.Remove(savedImagePath))
}
