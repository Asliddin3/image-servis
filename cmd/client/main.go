package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

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
	fmt.Println("Upload file start")
	time.Sleep(time.Second * 2)
	var wg sync.WaitGroup
	for i, val := range imagesStore {
		wg.Add(1)
		go func(wg sync.WaitGroup, i int) {
			UploadImage(i, imageService, val)
			defer wg.Done()
		}(wg, i)
	}
	fmt.Println("Upload file stop")
	fmt.Println("Download file start")
	time.Sleep(time.Second * 5)
	for i, val := range imagesStore {
		wg.Add(1)
		go func(wg sync.WaitGroup, i int) {
			DownloadImage(i, imageService, val)
			defer wg.Done()
		}(wg, i)
	}
	fmt.Println("Download file stop")
	fmt.Println("Get images start")
	time.Sleep(time.Second * 2)

	// downloadFolder := "./download"
	// for _, val := range imagesStore {
	// 	path := fmt.Sprintf("%s/%s", downloadFolder, val)
	// 	err = os.Remove(path)
	// 	if err != nil {
	// 		fmt.Println("error removeing file")
	// 		return
	// 	}
	// 	path = fmt.Sprintf("%s/%s", downloadFolder, val)
	// 	err = os.Remove(path)
	// 	if err != nil {
	// 		fmt.Println("error removeing file")
	// 		return
	// 	}
	// }
	// for i := 0; i < 1000; i++ {
	// 	wg.Add(1)
	// 	go func(wg sync.WaitGroup, i int) {
	// 		GetImages(i, imageService)
	// 		defer wg.Done()
	// 	}(wg, i)
	// }
	// wg.Wait()
	// fmt.Println("Get images stop")

}

func GetImages(i int, imageService image.ImageServiceClient) {
	fmt.Printf("getImages gorutine %d start\n", i)
	defer fmt.Printf("getImages gorutine %d stop\n", i)
	_, err := imageService.GetImages(context.Background(), &image.Empty{})
	if err != nil {
		return
	}
	fmt.Printf("get images from gorutine %d\n", i)

}

func DownloadImage(i int, imageService image.ImageServiceClient, fileName string) {
	fmt.Printf("download gorutine %d start\n", i)
	defer fmt.Printf("download gorutine %d stop\n", i)
	ImageFolder := "./download"
	stream, err := imageService.DownloadFile(context.Background(), &image.ImageInfo{
		FileName: fileName})
	fmt.Printf("download gorutine %d connect\n error %v", i, err)

	if err != nil {
		fmt.Println("error downloading", err)
		return
	}
	imageData := bytes.Buffer{}
	for {
		err := contextError(stream.Context())
		if err != nil {
			fmt.Printf("error in stream %v", err)
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
			fmt.Printf("cannot write data to file %v\n", err)
			break
		}
	}
	err = stream.CloseSend()
	if err != nil {
		fmt.Println("error sending msg", err)
		return
	}
	imagePath := fmt.Sprintf("%s/%s", ImageFolder, fileName)
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
	savedImagePath := fmt.Sprintf("%s/%s", ImageFolder, fileName)
	fmt.Println("saved image path", savedImagePath)
	_, err = os.Stat(savedImagePath)
	if err != nil {
		fmt.Println("file does't exists")
		return
	}

}

func UploadImage(i int, imageService image.ImageServiceClient, fileName string) {
	fmt.Printf("upload gorutine %d start\n", i)
	stream, err := imageService.UploadFile(context.Background())
	fmt.Printf("upload gorutine %d connect error %v\n", i, err)

	if err != nil {
		fmt.Println("upload file err", err)
		return
	}
	req := &image.UploadImageRequest{
		Request: &image.UploadImageRequest_Info{
			Info: &image.ImageInfo{
				FileName: fileName,
			},
		},
	}
	err = stream.Send(req)
	if err != nil {
		fmt.Println("send file err", err)
		return
	}
	ImageFolder := "./upload"
	imagePath := fmt.Sprintf("%s/%s", ImageFolder, fileName)
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
	fmt.Printf("upload gorutine %d stop error %v\n", i, err)
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
