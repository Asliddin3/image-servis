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

/*DiskImageStore this struct
have mutex for working with  disc
and imageFolder
*/
type DiskImageStore struct {
	mutex       sync.RWMutex
	imageFolder string
}

/*NewDiskImageStore this func used
for declaration imageFolder and mutex
*/
func NewDiskImageStore(imageFolder string) *DiskImageStore {
	return &DiskImageStore{
		imageFolder: imageFolder,
	}
}

/*Save This func save file in disc
 */
func (store *DiskImageStore) Save(
	fileName string,
	imageData bytes.Buffer,
) error {
	store.mutex.Lock()
	imagePath := fmt.Sprintf("%s/%s", store.imageFolder, fileName)
	fmt.Println(imagePath)
	file, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("cannot create image file: %w", err)
	}

	_, err = imageData.WriteTo(file)
	if err != nil {
		return fmt.Errorf("cannot write image to file: %w", err)
	}
	defer store.mutex.Unlock()

	return nil
}

/*GetImage This func get images from disc
and send to stream
*/
func (store *DiskImageStore) GetImage(
	FileName string,
	stream image.ImageService_DownloadFileServer,
) error {

	imagePath := fmt.Sprintf("%s/%s", store.imageFolder, FileName)

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
		return fmt.Errorf("error sending msg to stream %v", err)
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()

	return nil
}
