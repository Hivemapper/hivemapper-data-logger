package logger

import (
	"encoding/json"
	"fmt"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/imu-controller/device/iim42652"
	"os"
	"path"
	"sync"
	"time"
)

type DataWrapper struct {
	Time time.Time `json:"time"`
	Data any       `json:"data"`
}

func NewDataWrapper(time time.Time, data any) *DataWrapper {
	return &DataWrapper{
		Time: time,
		Data: data,
	}
}

type Accel struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func NewAccel(x, y, z float64) *Accel {
	return &Accel{
		X: x,
		Y: y,
		Z: z,
	}
}

type Gyro struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

func NewGyro(x, y, z float64) *Gyro {
	return &Gyro{
		X: x,
		Y: y,
		Z: z,
	}
}

type ImuDataWrapper struct {
	Accel *Accel    `json:"accel"`
	Gyro  *Gyro     `json:"gyro"`
	Temp  float64   `json:"temp"`
	Time  time.Time `json:"time"`
}

func NewImuDataWrapper(temperature iim42652.Temperature, acceleration *imu.Acceleration, angularRate *iim42652.AngularRate) *ImuDataWrapper {
	return &ImuDataWrapper{
		Accel: NewAccel(acceleration.X, acceleration.Y, acceleration.Z),
		Gyro:  NewGyro(angularRate.X, angularRate.Y, angularRate.Z),
		Time:  acceleration.Time,
		Temp:  *temperature,
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

	j.StartStoring()

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
