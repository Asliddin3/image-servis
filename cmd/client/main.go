package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Asliddin3/image-servis/config"
	"github.com/Asliddin3/image-servis/genproto/image"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", cfg.ImageServiceHost, cfg.ImageServicePort), grpc.WithInsecure())
	fmt.Println(fmt.Sprintf("%s:%s", cfg.ImageServiceHost, cfg.ImageServicePort))
	if err != nil {
		fmt.Println("dial error", err)
		return
	}
	imagesStore := []string{"laptop0.jpg", "gopher.jpg",
		"gopher1.png", "gopher2.jpeg", "gopher3.jpeg", "gopher4.png",
		"gopher5.png", "laptop1.jpeg", "laptop2.jpeg", "laptop3.jpeg",
		"laptop4.jpeg", "laptop5.jpeg",
	}
	imageService := image.NewImageServiceClient(conn)
	uploadCh := make(chan struct{}, 10)
	for _, val := range imagesStore {
		uploadCh <- struct{}{}
		go UploadImage(uploadCh, imageService, val)
	}

}

func UploadImage(ch chan struct{}, imageService image.ImageServiceClient, fileName string) {
	stream, err := imageService.UploadFile(context.Background())
	if err != nil {
		fmt.Println("upload file err", err)
		return
	}
	imageType := filepath.Ext(fileName)
	fmt.Println(imageType)
	imageName := strings.TrimRight(fileName, imageType)
	fmt.Println(imageName)
	req := &image.UploadImageRequest{
		Request: &image.UploadImageRequest_Info{
			Info: &image.ImageInfo{
				ImageName: imageName,
				ImageData: imageType,
			},
		},
	}
	err = stream.Send(req)
	if err != nil {
		fmt.Println("send file err", err)
		return
	}
	ImageFolder := "./tmp"
	imagePath := fmt.Sprintf("%s/%s%s", ImageFolder, imageName, imageType)
	fmt.Println(imagePath)
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Println("error opening file", err)
		return
	}
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			fmt.Println("EOF in for")
			break
		}
		if err != nil {
			fmt.Println("reading file err", err)
			return
		}

		req := &image.UploadImageRequest{
			Request: &image.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}
		err = stream.Send(req)
		if err != nil {
			fmt.Println("stream send byte err", err)
			return
		}
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		fmt.Println("close stream ", err)
		return
	}
	<-ch
}
