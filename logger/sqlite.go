package logger

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/Hivemapper/gnss-controller/device/neom9n"
	"github.com/Hivemapper/hivemapper-data-logger/data/imu"
	"github.com/Hivemapper/hivemapper-data-logger/data/session"
	"github.com/streamingfast/imu-controller/device/iim42652"

	_ "modernc.org/sqlite"
)

type Sqlite struct {
	lock                     sync.Mutex
	DB                       *sql.DB
	file                     string
	doInsert                 bool
	purgeQueryFuncList       []PurgeQueryFunc
	createTableQueryFuncList []CreateTableQueryFunc
	alterTableQueryFuncList  []AlterTableQueryFunc

	logs chan Sqlable
}

func (s *Sqlite) GetConfig(key string) string {
	row := s.DB.QueryRow("SELECT value FROM config WHERE key = ?", key)
	var result string
	err := row.Scan(&result)
	if err != nil {
		result = ""
	}
	return result
}

func NewSqlite(file string, createTableQueryFuncList []CreateTableQueryFunc, alterTableQueryFuncList []AlterTableQueryFunc, purgeQueryFuncList []PurgeQueryFunc) *Sqlite {
	return &Sqlite{
		file:                     file,
		createTableQueryFuncList: createTableQueryFuncList,
		alterTableQueryFuncList:  alterTableQueryFuncList,
		purgeQueryFuncList:       purgeQueryFuncList,
		logs:                     make(chan Sqlable, 1000),
	}
}

func (s *Sqlite) InsertErrorLog(message string) {
	s.DB.Exec("INSERT OR IGNORE INTO error_logs VALUES (NULL,?,?,?)", time.Now().Format("2006-01-02 15:04:05.99999"), "data-logger", message)
}

func dumpDatabase(dbFilePath, dumpFilePath string) error {
	cmd := exec.Command("sqlite3", dbFilePath, ".recover")
	output, err := os.Create(dumpFilePath)
	if err != nil {
		return err
	}
	defer output.Close()

	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func rebuildDatabase(dumpFilePath, newDbFilePath string) error {
	// Remove the new database file if it already exists
	if _, err := os.Stat(newDbFilePath); err == nil {
		err = os.Remove(newDbFilePath)
		if err != nil {
			return fmt.Errorf("failed to remove existing new database file: %v", err)
		}
	}

	cmd := exec.Command("sqlite3", newDbFilePath)
	input, err := os.Open(dumpFilePath)
	if err != nil {
		return err
	}
	defer input.Close()

	cmd.Stdin = input
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func recoverDb(corruptedDB string) {
	dumpFile := corruptedDB + ".dump"
	newDB := corruptedDB + ".recovered"

	// Dump corrupted database
	err := dumpDatabase(corruptedDB, dumpFile)
	if err != nil {
		log.Fatalf("Failed to dump database: %v", err)
	}
	log.Println("Database dumped successfully.")

	// Rebuild database
	err = rebuildDatabase(dumpFile, newDB)
	if err != nil {
		log.Fatalf("Failed to rebuild database: %v", err)
	}
	log.Println("Database rebuilt successfully.")

	log.Println("Removing corrupted database and renaming new database.")
	// Remove corrupted database
	err = os.Remove(corruptedDB)
	err = os.Remove(corruptedDB + "-shm")
	err = os.Remove(corruptedDB + "-wal")

	// Rename new database to the original name
	err = os.Rename(newDB, corruptedDB)

	// Remove dump file
	err = os.Remove(dumpFile)
}

func (s *Sqlite) Init(logTTL time.Duration) error {
	fmt.Println("initializing database:", s.file)
	db, err := sql.Open("sqlite", s.file)

	if err != nil {
		return fmt.Errorf("opening database: %s", err.Error())
	}

	// check integrity of the database
	shouldRecover := false
	row := db.QueryRow("PRAGMA integrity_check;")
	var result string
	if err := row.Scan(&result); err != nil {
		fmt.Printf("failed to perform integrity check: %v\n", err)
		shouldRecover = true
	}
	fmt.Println("Integrity Check Result:", result)
	if result != "ok" {
		shouldRecover = true
	}

	if shouldRecover {
		fmt.Println("Recovering database and stopping the program.")
		if err := db.Close(); err != nil {
			fmt.Printf("failed to close database: %v\n", err)
		}
		recoverDb(s.file)
		os.Exit(1)
	}

	// Enable Write-Ahead Logging
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return fmt.Errorf("setting WAL mode: %s", err.Error())
	}

	_, err = db.Exec("PRAGMA busy_timeout = 1000;")
	if err != nil {
		return fmt.Errorf("setting BUSY timeout: %s", err.Error())
	}

	for _, createQuery := range s.createTableQueryFuncList {
		if _, err := db.Exec(createQuery()); err != nil {
			return fmt.Errorf("creating table: %s", err.Error())
		}
	}

	for _, alterQuery := range s.alterTableQueryFuncList {
		if _, err := db.Exec(alterQuery()); err != nil {
			fmt.Println("field exists:", err.Error())
		}
	}

	fmt.Println("WAL checkpoint 1")
	_, err = db.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
	if err != nil {
		return fmt.Errorf("checkpoint: %s", err.Error())
	}

	fmt.Println("Vacuum DB")
	_, err = db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("VACUUM: %s", err.Error())
	}

	fmt.Println("WAL checkpoint 2")
	_, err = db.Exec("PRAGMA wal_checkpoint(TRUNCATE);")
	if err != nil {
		return fmt.Errorf("checkpoint: %s", err.Error())
	}

	fmt.Println("database initialized, will purge every:", logTTL.String())

	s.DB = db

	err = s.initiateSession()
	if err != nil {
		return fmt.Errorf("init session: %s", err.Error())
	}

	if logTTL > 0 {
		go func() {
			for {
				time.Sleep(5 * time.Minute)
				s.InsertErrorLog("purging database")
				err := s.Purge(logTTL)
				if err != nil {
					s.InsertErrorLog("purging database error: " + err.Error())
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
		hardRetries := 0
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
			// Properly format and accumulate fields and parameters
			if len(fields) > 0 {
				// Trim any trailing comma from cumulatedFields
				if len(accumulator.cumulatedFields) > 0 && accumulator.cumulatedFields[len(accumulator.cumulatedFields)-1] == ',' {
					accumulator.cumulatedFields = accumulator.cumulatedFields[:len(accumulator.cumulatedFields)-1]
				}
				// Append comma only if there are already fields
				if len(accumulator.cumulatedFields) > 0 {
					accumulator.cumulatedFields += ","
				}
				accumulator.cumulatedFields += fields
				accumulator.cumulatedParams = append(accumulator.cumulatedParams, params...)
			}

			if accumulator.count < log.BufferSize() {
				continue
			}

			if accumulator.cumulatedFields[len(accumulator.cumulatedFields)-1] == ',' {
				accumulator.cumulatedFields = accumulator.cumulatedFields[0 : len(accumulator.cumulatedFields)-1] //remove last comma
			}
			stmt, err := db.Prepare(query + accumulator.cumulatedFields)
			if err != nil {
				s.InsertErrorLog(err.Error())
				delete(queries, query)
			} else {
				maxRetries := 6 // Maximum number of retries
				for attempt := 0; attempt < maxRetries; attempt++ {
					_, err = stmt.Exec(accumulator.cumulatedParams...)
					if err == nil {
						delete(queries, query)
						break // Success, exit the retry loop
					}

					if attempt < maxRetries-1 {
						delay := time.Duration(100*(1<<attempt)) * time.Millisecond
						time.Sleep(delay)
						fmt.Println(err)
						s.InsertErrorLog(fmt.Sprintf("Retry attempt %d for query: %s", attempt+1, err))
					}
				}

				if err != nil {
					s.InsertErrorLog(err.Error()) // Log final failure
					fmt.Println(err)
					hardRetries++
					if hardRetries > 3 {
						delete(queries, query)
						hardRetries = 0
					}
				}
			}
		}
	}()

	return nil
}

func (s *Sqlite) initiateSession() error {
	// Assume s.DB is *sql.DB
	var lastImuTime sql.NullTime // Using sql.NullTime to handle possible null values
	err := s.DB.QueryRow("SELECT time FROM imu ORDER BY id DESC LIMIT 1").Scan(&lastImuTime)
	fmt.Println("last imu time:", lastImuTime.Time, time.Now(), lastImuTime.Time.Before(time.Now()))
	if err != nil {
		if err == sql.ErrNoRows {
			// No rows in the table, so generate new session
			err := session.SetSession("")
			if err != nil {
				return fmt.Errorf("setting session: %s", err.Error())
			}
		} else {
			return fmt.Errorf("query last imu time: %s", err.Error())
		}
	} else if lastImuTime.Valid && lastImuTime.Time.Before(time.Now()) {
		// If last imu time is valid and before the current time, then get session from last imu
		var sessionID string
		err = s.DB.QueryRow("SELECT session FROM imu ORDER BY id DESC LIMIT 1").Scan(&sessionID)
		if err != nil {
			return fmt.Errorf("getting session: %s", err.Error())
		}
		fmt.Println("Using session:", sessionID)
		err = session.SetSession(sessionID)
		if err != nil {
			return fmt.Errorf("setting session: %s", err.Error())
		}
	} else {
		// Generate new session
		err := session.SetSession("")
		if err != nil {
			return fmt.Errorf("setting session: %s", err.Error())
		}
	}
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
	start := time.Now()
	for _, purgeQueryFunc := range s.purgeQueryFuncList {
		if res, err := s.DB.Exec(purgeQueryFunc(), t); err != nil {
			return err
		} else {
			c, _ := res.RowsAffected()
			fmt.Println("purged rows:", c, "in", time.Since(start).String())
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
