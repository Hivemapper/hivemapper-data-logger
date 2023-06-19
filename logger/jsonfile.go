package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

type DataWrapper struct {
	Time time.Time `json:"time"`
	Data any       `json:"Data"`
}

func NewDataWrapper(time time.Time, data any) *DataWrapper {
	return &DataWrapper{
		Time: time,
		Data: data,
	}
}

type JsonFile struct {
	datas        []*DataWrapper
	lock         sync.Mutex
	destFolder   string
	saveInterval time.Duration
}

func NewJsonFile(destFolder string, saveInterval time.Duration) *JsonFile {
	fmt.Println("creating json file logger:", destFolder, "save interval:", saveInterval.String())
	return &JsonFile{
		saveInterval: saveInterval,
		destFolder:   destFolder,
	}
}

func (j *JsonFile) Init() error {
	fmt.Println("initializing json file logger")
	latestLog := path.Join(j.destFolder, "latest.log")
	if fileExists(latestLog) {
		err := os.Remove(latestLog)
		if err != nil {
			return fmt.Errorf("removing latest.log: %w", err)
		}
		fmt.Println("removed:", latestLog)
	}

	if !fileExists(j.destFolder) {
		fmt.Printf("Creating destination folder: %s\n", j.destFolder)
		err := os.MkdirAll(j.destFolder, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating destination folder %s : %w", j.destFolder, err)
		}
	}

	return nil
}

func (j *JsonFile) StartStoring() {
	go func() {
		for {
			fmt.Println("saving to file with entry count:", len(j.datas))
			if len(j.datas) > 0 {
				err := j.toFile(j.datas[0].Time)
				if err != nil {
					panic(fmt.Errorf("writing to file: %w", err))
				}
				j.datas = nil
			}
			time.Sleep(j.saveInterval)
		}
	}()
}

func (j *JsonFile) Log(time time.Time, data any) error {
	j.lock.Lock()
	defer j.lock.Unlock()
	j.datas = append(j.datas, NewDataWrapper(time, data))
	err := writeToFile(path.Join(j.destFolder, "latest.log"), data)
	if err != nil {
		return fmt.Errorf("writing latest file: %w", err)
	}
	return nil
}

func (j *JsonFile) toFile(time time.Time) error {
	j.lock.Lock()
	defer j.lock.Unlock()
	if len(j.datas) == 0 {
		return nil
	}

	fileName := fmt.Sprintf("%s.json", time.Format("2006-01-02T15:04:05.000Z"))
	filePath := path.Join(j.destFolder, fileName)

	err := writeAllToFile(filePath, j.datas)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

func writeAllToFile(filePath string, dw []*DataWrapper) error {
	jsonData, err := json.Marshal(dw)
	if err != nil {
		return fmt.Errorf("marshaling Data: %w", err)
	}
	err = os.WriteFile(filePath, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing file '%s': %w", filePath, err)
	}

	return nil
}

func writeToFile(filePath string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling Data: %w", err)
	}
	err = os.WriteFile(filePath, jsonData, os.ModePerm)
	if err != nil {
		return fmt.Errorf("writing file '%s': %w", filePath, err)
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
