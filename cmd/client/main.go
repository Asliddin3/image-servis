package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Asliddin3/image-servis/config"
	"github.com/Asliddin3/image-servis/genproto/image"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	var wg sync.WaitGroup
	for i, val := range imagesStore {
		wg.Add(1)
		go UploadImage(i, imageService, val, wg)
	}
	wg.Wait()
	for i, val := range imagesStore {
		wg.Add(1)
		go DownloadImage(i, imageService, val, wg)
	}
	wg.Wait()
	downloadFolder := "../client/download"
	imgFolder := "../../img"
	for _, val := range imagesStore {
		path := fmt.Sprintf("%s/%s", downloadFolder, val)
		err = os.Remove(path)
		if err != nil {
			fmt.Println("error removeing file")
			return
		}
		path = fmt.Sprintf("%s/%s", imgFolder, val)
		err = os.Remove(path)
		if err != nil {
			fmt.Println("error removeing file")
			return
		}
	}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go GetImages(i, imageService, wg)
	}
	wg.Wait()

}

func GetImages(i int, imageService image.ImageServiceClient, wg sync.WaitGroup) {
	fmt.Printf("getImages gorutine %d start", i)
	defer fmt.Printf("getImages gorutine %d stop", i)
	defer wg.Done()
	res, err := imageService.GetImages(context.Background(), &image.Empty{})
	if err != nil {
		return
	}
	fmt.Println(res)
}

func DownloadImage(i int, imageService image.ImageServiceClient, fileName string, wg sync.WaitGroup) {
	fmt.Printf("download gorutine %d start", i)
	defer fmt.Printf("download gorutine %d stop", i)
	defer wg.Done()
	ImageFolder := "../download"
	imageType := filepath.Ext(fileName)
	imageName := strings.TrimRight(fileName, imageType)
	stream, err := imageService.DownloadFile(context.Background(), &image.ImageInfo{
		ImageName: imageName,
		ImageData: imageType})
	if err != nil {
		fmt.Println("error downloading", err)
		return
	}
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
		if err != nil {
			fmt.Printf("cannot receive chunk data %s\n", err.Error())
			break
		}
		chunk := req.GetChunkData()
		_, err = imageData.Write(chunk)
		if err != nil {
			break
		}
	}
	err = stream.CloseSend()
	if err != nil {
		fmt.Println("error sending msg", err)
		return
	}

	imagePath := fmt.Sprintf("%s/%s%s", ImageFolder, imageName, imageType)
	fmt.Println(imagePath)
	file, err := os.Create(imagePath)
	if err != nil {
		fmt.Println("eror creating file", err)
		return
	}
	defer file.Close()

	_, err = imageData.WriteTo(file)
	if err != nil {
		fmt.Println("error writing tp file")
		return
	}
	savedImagePath := fmt.Sprintf("%s/%s%s", ImageFolder, imageName, imageType)
	fmt.Println("saved image path", savedImagePath)
	_, err = os.Stat(savedImagePath)
	if err != nil {
		fmt.Println("file does't exists")
		return
	}

}

func UploadImage(i int, imageService image.ImageServiceClient, fileName string, wg sync.WaitGroup) {
	fmt.Printf("upload gorutine %d start", i)
	defer fmt.Printf("upload gorutine %d stop", i)
	defer wg.Done()
	stream, err := imageService.UploadFile(context.Background())
	if err != nil {
		fmt.Println("upload file err", err)
		return
	}
	imageType := filepath.Ext(fileName)
	imageName := strings.TrimRight(fileName, imageType)
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
	ImageFolder := "../upload"
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
		fmt.Println("cannot close stream ", err)
		return
	}

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
