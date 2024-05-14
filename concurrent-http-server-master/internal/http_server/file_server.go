package httpserver

import (
	"concurrent/pkg"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	CHUNCK_SIZE int64 = 64
	BASE_PATH         = "../files"
)

type FileServer struct {
	fileDb pkg.FileDB
	logger *logrus.Logger
}

func NewFileServer(fileDb pkg.FileDB, logger *logrus.Logger) *FileServer {
	return &FileServer{
		fileDb: fileDb,
		logger: logger,
	}
}

func getNextDirID(basePath string) (int, error) {
	dirRegex := regexp.MustCompile(`dir_(\d+)`)

	files, err := ioutil.ReadDir(basePath)

	if err != nil {
		return 0, err
	}

	maxID := 0
	for _, f := range files {
		if f.IsDir() {
			matches := dirRegex.FindStringSubmatch(f.Name())
			if matches != nil {
				id, _ := strconv.Atoi(matches[1])
				if id > maxID {
					maxID = id
				}
			}
		}
	}

	return maxID + 1, nil
}

func createNewFile(basePath, fileName string) (*os.File, string, error) {
	nextID, err := getNextDirID(basePath)
	if err != nil {
		return nil, "", err
	}

	newDirPath := filepath.Join(basePath, fmt.Sprintf("dir_%d", nextID))
	if err := os.Mkdir(newDirPath, 0755); err != nil {
		return nil, "", err
	}

	filePath := filepath.Join(newDirPath, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, "", err
	}

	return file, filePath, nil
}

func (fs *FileServer) WriteToFile(content []byte, fileName string) (string, error) {
	var mtx = &sync.Mutex{}
	var wg = &sync.WaitGroup{}

	file, filePath, err := createNewFile(BASE_PATH, fileName)
	if err != nil {
		return "", err
	}
	var size int64
	size = int64(len(content))

	numChunks := size / CHUNCK_SIZE
	if size%CHUNCK_SIZE != 0 {
		numChunks++
	}
	wg.Add(int(numChunks))

	writeToFile := func(cnt []byte, offset int64) {
		defer wg.Done()
		end := offset + CHUNCK_SIZE
		if end > size {
			end = size
		}
		mtx.Lock()

		if _, err := file.WriteAt(cnt[offset:end], offset); err != nil {
			fs.logger.WithError(err).Warn("Error writing to file:")
		}
		mtx.Unlock()
	}
	var start int64
	for start = 0; start < size; start += CHUNCK_SIZE {
		go writeToFile(content, start)
	}
	wg.Wait()
	file.Close()
	//	Make Access Hash
	var num uint64
	err = binary.Read(rand.Reader, binary.BigEndian, &num)
	if err != nil {
		return "", err
	}

	sEnc := base64.StdEncoding.EncodeToString([]byte(fileName))
	fileId := strconv.FormatUint(num, 10) + ":" + sEnc

	err = fs.fileDb.AddToDB(fileId, filePath)
	if err != nil {
		return "", err
	}

	return fileId, nil
}

func (fs *FileServer) ReadFromFile(fileId string) ([]byte, string, error) {
	//var rwMutex = &sync.RWMutex{}
	var wg = &sync.WaitGroup{}
	filePath, err := fs.fileDb.GetFilePath(fileId)
	if err != nil {
		return nil, "", err
	}

	//	Get Size of file
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, "", err
	}
	fileSize := fileInfo.Size()

	// Open file once
	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	//	Extract file name
	fileName := file.Name()
	pathParts := strings.Split(fileName, "/")
	fileName = pathParts[len(pathParts)-1]

	content := make([]byte, fileSize)
	//	Read it concurrent
	readFromFile := func(offset int64) {
		defer wg.Done()
		end := offset + CHUNCK_SIZE
		if end > fileSize {
			end = fileSize
		}
		cnt := make([]byte, end-offset)
		if _, err := file.ReadAt(cnt, offset); err != nil {
			fs.logger.WithError(err).Warn("Error reading from file:")
		}
		copy(content[offset:], cnt)
	}

	numChunks := int(fileSize / CHUNCK_SIZE)
	if fileSize%CHUNCK_SIZE != 0 {
		numChunks++
	}
	wg.Add(numChunks)

	var start int64
	for start = 0; start < fileSize; start += CHUNCK_SIZE {
		go readFromFile(start)
	}
	wg.Wait()
	return content, fileName, nil
}
