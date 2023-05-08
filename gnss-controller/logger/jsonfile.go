package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

type FileInfo struct {
	path             string
	size             int64
	modificationTime time.Time
}

type JsonFile struct {
	data          *Data
	lock          sync.Mutex
	destFolder    string
	maxFolderSize int64
	currentSize   int64
	maxSize       int64
	files         []*FileInfo
	saveInterval  time.Duration
}

func NewJsonFile(destFolder string, maxFolderSize int64, saveInterval time.Duration) *JsonFile {
	return &JsonFile{
		saveInterval:  saveInterval,
		destFolder:    destFolder,
		maxFolderSize: maxFolderSize,
	}
}

func (j *JsonFile) Init() error {
	latestLog := path.Join(j.destFolder, "latest.log")
	if fileExists(latestLog) {
		err := os.Remove(latestLog)
		if err != nil {
			return fmt.Errorf("removing latest.log: %w", err)
		}
		fmt.Println("removed:", latestLog)
		return nil
	}

	if !fileExists(j.destFolder) {
		fmt.Printf("Creating destination folder: %s\n", j.destFolder)
		err := os.MkdirAll(j.destFolder, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating destination folder %s : %w", j.destFolder, err)
		}
	}

	files, err := os.ReadDir(j.destFolder)
	if err != nil {
		panic(fmt.Errorf("listing destionation folder %s : %w", j.destFolder, err))
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		i, err := f.Info()
		if err != nil {
			return fmt.Errorf("getting file info: %s , %w", f.Name(), err)
		}

		if f.Name() == "latest.log" {
			continue
		}

		fi := &FileInfo{
			path:             path.Join(j.destFolder, f.Name()),
			size:             i.Size(),
			modificationTime: i.ModTime(),
		}
		j.addFile(fi)
	}

	return nil
}

func (j *JsonFile) startStoring() {
	go func() {
		for {
			time.Sleep(j.saveInterval)
			err := j.toFile()
			if err != nil {
				panic(fmt.Errorf("writing to file: %w", err))
			}
		}
	}()
}

func (j *JsonFile) addFile(f *FileInfo) {
	j.files = append(j.files, f)
	j.currentSize += f.size

}
func (j *JsonFile) Log(data *Data) error {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.data = data
	err := writeToFile(path.Join(j.destFolder, "latest.log"), j.data)
	if err != nil {
		return fmt.Errorf("writing latest file: %w", err)
	}
	return nil
}

func (j *JsonFile) toFile() error {
	j.lock.Lock()
	defer j.lock.Unlock()
	if j.data == nil {
		return nil
	}

	fileName := fmt.Sprintf("%s.json", j.data.Timestamp.Format("2006-01-02T15:04:05.000Z"))
	filePath := path.Join(j.destFolder, fileName)

	err := writeToFile(filePath, j.data)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	fi, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("getting file info: %w", err)
	}
	j.addFile(&FileInfo{
		path:             filePath,
		size:             fi.Size(),
		modificationTime: fi.ModTime(),
	})
	err = j.freeUpSpace(fi.Size())
	if err != nil {
		return fmt.Errorf("freeing up space: %w", err)
	}
	return nil
}

func writeToFile(filePath string, data *Data) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling data: %w", err)
	}
	err = os.WriteFile(filePath, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing file '%s': %w", filePath, err)
	}

	return nil
}

func (j *JsonFile) freeUpSpace(nextFileSize int64) error {
	if j.currentSize+nextFileSize > j.maxSize {
		spaceToReclaim := j.maxSize * 10 / 100
		for spaceToReclaim > 0 {
			fi := j.files[0]
			j.files = j.files[1:]

			if fileExists(fi.path) {
				err := os.Remove(fi.path)
				if err != nil {
					return fmt.Errorf("removing file: %s, %w", fi.path, err)
				}
			} else {
				log.Println("free space: skipping file that does not exist anymore: ", fi.path)
			}

			j.currentSize -= fi.size
			spaceToReclaim -= fi.size
		}
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}
