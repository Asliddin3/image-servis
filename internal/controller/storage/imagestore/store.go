package imagestore

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/Asliddin3/image-servis/genproto/image"
)

type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
	Images      map[string]*ImageInfo
}
type ImageInfo struct {
	Name string
	Type string
	Path string
}

func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
		Images:      make(map[string]*ImageInfo),
	}
}

func (store *DiskImageStore) Save(
	imageName string,
	imageType string,
	imageData bytes.Buffer,
) error {

	imagePath := fmt.Sprintf("%s/%s%s", store.imageFolder, imageName, imageType)
	fmt.Println(imagePath)
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("cannot create image file: %w", err)
	}

	_, err = imageData.WriteTo(file)
	if err != nil {
		return fmt.Errorf("cannot write image to file: %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.Images[imageName] = &ImageInfo{
		Name: imageName,
		Type: imageType,
		Path: imagePath,
	}
	return nil
}

func (store *DiskImageStore) GetImage(
	ImageName string,
	ImageType string,
	stream image.ImageService_DownloadFileServer,
) error {

	imagePath := fmt.Sprintf("%s/%s.%s", store.imageFolder, ImageName, ImageType)

	file, err := os.Open(imagePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("cannot open image file: %w", err)
	}
	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		size += n

		req := &image.ImageResponse{
			ChunkData: buffer[:n],
		}

		err = stream.Send(req)
	}
	var m interface{}
	err = stream.SendMsg(m)
	if err != nil {
		return fmt.Errorf("error sending msg to stream", err)
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()

	return nil
}
