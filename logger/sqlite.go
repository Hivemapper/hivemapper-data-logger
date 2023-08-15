package logger

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"github.com/streamingfast/gnss-controller/device/neom9n"
	"github.com/streamingfast/hivemapper-data-logger/data/imu"
	"github.com/streamingfast/imu-controller/device/iim42652"
	"io"
	"os"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type Sqlite struct {
	lock                     sync.Mutex
	DB                       *sql.DB
	file                     string
	doInsert                 bool
	purgeQueryFuncList       []PurgeQueryFunc
	createTableQueryFuncList []CreateTableQueryFunc

	logs chan Sqlable
}

func NewSqlite(file string, createTableQueryFuncList []CreateTableQueryFunc, purgeQueryFuncList []PurgeQueryFunc) *Sqlite {
	return &Sqlite{
		file:                     file,
		createTableQueryFuncList: createTableQueryFuncList,
		purgeQueryFuncList:       purgeQueryFuncList,
		logs:                     make(chan Sqlable, 200),
	}
}

func (s *Sqlite) Init(logTTL time.Duration) error {
	fmt.Println("initializing database:", s.file)
	db, err := sql.Open("sqlite", s.file)

	if err != nil {
		return fmt.Errorf("opening database: %s", err.Error())
	}

	for _, createQuery := range s.createTableQueryFuncList {
		if _, err := db.Exec(createQuery()); err != nil {
			return fmt.Errorf("creating table: %s", err.Error())
		}
	}

	fmt.Println("database initialized, will purge every:", logTTL.String())

	if logTTL > 0 {
		go func() {
			for {
				time.Sleep(time.Minute)
				err := s.Purge(logTTL)
				if err != nil {
					panic(fmt.Errorf("purging database: %s", err.Error()))
				}
			}
		}()
	}

	go func() {
		type Accumulator struct {
			count           int
			cumulatedParams []any
			cumulatedFields string
		}
		queries := map[string]*Accumulator{}
		for {
			log := <-s.logs
			query, fields, params := log.InsertQuery()

			if query == "" {
				continue
			}

			accumulator, found := queries[query]
			if !found {
				accumulator = &Accumulator{}
				queries[query] = accumulator
			}
			accumulator.count++
			accumulator.cumulatedFields += fields
			accumulator.cumulatedParams = append(accumulator.cumulatedParams, params...)

			if accumulator.count < 100 {
				continue
			}

			accumulator.cumulatedFields = accumulator.cumulatedFields[0 : len(accumulator.cumulatedFields)-1] //remove last comma
			stmt, err := db.Prepare(query + accumulator.cumulatedFields)
			if err != nil {
				panic(fmt.Errorf("preparing statement for inserting Data: %w", err))
			}
			s.lock.Lock()
			_, err = stmt.Exec(accumulator.cumulatedParams...)
			s.lock.Unlock()
			if err != nil {
				panic(fmt.Errorf("inserting Data: %s", err.Error()))
			}
			delete(queries, query)
		}
	}()

	s.DB = db

	return nil
}

func (s *Sqlite) Clone() (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	cloneFilename := fmt.Sprintf("%s_clone.db.gz", s.file)

	sourceFileStat, err := os.Stat(s.file)
	if err != nil {
		return "", fmt.Errorf("database does not exist: %w", err)
	}

	if !sourceFileStat.Mode().IsRegular() {
		return "", fmt.Errorf("%s is not a regular file", s.file)
	}

	source, err := os.Open(s.file)
	if err != nil {
		return "", fmt.Errorf("opening database: %w", err)
	}
	defer source.Close()

	gzippedFile, err := os.Create(cloneFilename)
	if err != nil {
		panic(err)
	}
	defer gzippedFile.Close()

	gzipWriter := gzip.NewWriter(gzippedFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, source)
	if err != nil {
		return "", fmt.Errorf("copying zipped database: %w", err)
	}

	err = gzipWriter.Flush()
	if err != nil {
		return "", fmt.Errorf("flushing gzip writer: %w", err)
	}

	return cloneFilename, nil
}

func (s *Sqlite) FetchRawMergedData(from string, to string, includeImu bool, includeGnss bool) ([]*JsonDataWrapper, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	rows, err := s.DB.Query("SELECT * FROM merged WHERE imu_time > ? AND imu_time < ?", from, to)
	if err != nil {
		return nil, fmt.Errorf("querying last position: %s", err.Error())
	}

	if err != nil {
		return nil, fmt.Errorf("querying last position: %s", err.Error())
	}
	defer rows.Close()

	var jsonData []*JsonDataWrapper

	for rows.Next() {
		id := 0
		imuTime := time.Time{}
		temperature := iim42652.NewTemperature(0.0)
		acceleration := &iim42652.Acceleration{}
		gnssData := &neom9n.Data{
			SystemTime: time.Time{},
			Timestamp:  time.Time{},
			Dop:        &neom9n.Dop{},
			Satellites: &neom9n.Satellites{},
			RF:         &neom9n.RF{},
		}
		gyro := &Gyro{
			X: 0,
			Y: 0,
			Z: 0,
		}
		var _camOrientation *imu.Orientation // read the data, but do not use it in the json writer, not useful as of now
		err := rows.Scan(
			&id,
			&imuTime,
			&acceleration.TotalMagnitude,
			&acceleration.X,
			&gyro.X,
			&acceleration.Y,
			&gyro.Y,
			&acceleration.Z,
			&gyro.Z,
			&temperature,
			&_camOrientation,
			&gnssData.SystemTime,
			&gnssData.Timestamp,
			&gnssData.Fix,
			&gnssData.Ttff,
			&gnssData.Latitude,
			&gnssData.Longitude,
			&gnssData.Altitude,
			&gnssData.Speed,
			&gnssData.Heading,
			&gnssData.Satellites.Seen,
			&gnssData.Satellites.Used,
			&gnssData.Eph,
			&gnssData.HorizontalAccuracy,
			&gnssData.VerticalAccuracy,
			&gnssData.HeadingAccuracy,
			&gnssData.SpeedAccuracy,
			&gnssData.Dop.HDop,
			&gnssData.Dop.VDop,
			&gnssData.Dop.XDop,
			&gnssData.Dop.YDop,
			&gnssData.Dop.TDop,
			&gnssData.Dop.PDop,
			&gnssData.Dop.GDop,
			&gnssData.RF.JammingState,
			&gnssData.RF.AntStatus,
			&gnssData.RF.AntPower,
			&gnssData.RF.PostStatus,
			&gnssData.RF.NoisePerMS,
			&gnssData.RF.AgcCnt,
			&gnssData.RF.JamInd,
			&gnssData.RF.OfsI,
			&gnssData.RF.MagI,
			&gnssData.RF.OfsQ,
		)

		jsonDataWrapper := NewJsonDataWrapper(nil, nil, nil, gyro)

		if includeImu {
			jsonDataWrapper.Acceleration = imu.NewAcceleration(acceleration.X, acceleration.Y, acceleration.Z, acceleration.TotalMagnitude, imuTime)
			jsonDataWrapper.Temperature = temperature
		}

		if includeGnss {
			jsonDataWrapper.GnssData = gnssData
		}

		jsonData = append(jsonData, jsonDataWrapper)

		if err != nil {
			return nil, fmt.Errorf("scanning last position: %s", err.Error())
		}
	}

	return jsonData, nil
}

func (s *Sqlite) Purge(ttl time.Duration) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	t := time.Now().Add(ttl * -1)
	fmt.Println("purging database older than:", t)
	for _, purgeQueryFunc := range s.purgeQueryFuncList {
		if res, err := s.DB.Exec(purgeQueryFunc(), t); err != nil {
			return err
		} else {
			c, _ := res.RowsAffected()
			fmt.Println("purged rows:", c)
		}
	}

	return nil
}

func (s *Sqlite) Log(data Sqlable) error {
	s.logs <- data
	return nil
}

func (s *Sqlite) SingleRowQuery(sql string, handleRow func(row *sql.Rows) error, params ...any) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	rows, err := s.DB.Query(sql, params...)
	if err != nil {
		return fmt.Errorf("querying last position: %s", err.Error())
	}
	defer rows.Close()

	if rows.Next() {
		err := handleRow(rows)
		if err != nil {
			return fmt.Errorf("handling row: %s", err.Error())
		}
		return nil
	}

	return nil
}

func (s *Sqlite) Query(debugLogQuery bool, sql string, handleRow func(row *sql.Rows) error, params []any) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if debugLogQuery {
		fmt.Println("Running query:", sql, params)
	}

	rows, err := s.DB.Query(sql, params...)
	if err != nil {
		return fmt.Errorf("querying last position: %s", err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		err := handleRow(rows)
		if err != nil {
			return fmt.Errorf("handling row: %s", err.Error())
		}
	}

	return nil
}
