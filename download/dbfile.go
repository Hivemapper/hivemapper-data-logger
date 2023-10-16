package download

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/Hivemapper/hivemapper-data-logger/logger"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type Download struct {
	sqliteLogger *logger.Sqlite
}

func NewDownload(sqliteLogger *logger.Sqlite) *Download {
	return &Download{sqliteLogger: sqliteLogger}
}

func (d *Download) GetRawData(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	if from == "" {
		fmt.Fprintf(w, "error: missing 'from' query parameter\n")
		return
	}

	to := r.URL.Query().Get("to")
	if to == "" {
		fmt.Fprintf(w, "error: missing 'to' query parameter\n")
		return
	}

	includeImu := true
	includeGnss := true

	includeImuParam := r.URL.Query().Get("includeImu")
	includeGnssParam := r.URL.Query().Get("includeGnss")

	if includeImuParam == "false" && includeGnssParam == "false" {
		fmt.Fprintf(w, "no data requested\n")
		return
	}

	if includeImuParam == "false" {
		includeImu = false
	}

	if includeGnssParam == "false" {
		includeGnss = false
	}

	jsonData, err := d.sqliteLogger.FetchRawMergedData(from, to, includeImu, includeGnss)
	if err != nil {
		fmt.Fprintf(w, "fetching raw merged data: %s", err)
		return
	}

	data, err := json.Marshal(jsonData)
	if err != nil {
		fmt.Fprintf(w, "marshalling json: %s", err)
		return
	}
	fmt.Println("Done marshalling json", len(data))

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err = gz.Write(data)
	if err != nil {
		fmt.Fprintf(w, "compressing data: %s", err)
		return
	}
	gz.Close()

	w.Header().Set("Content-Type", "application/x-gzip")
	w.Write(buf.Bytes())
	return
}

func (d *Download) GetDatabaseFiles(w http.ResponseWriter, _ *http.Request) {
	fmt.Println("Cloning and compressing the database...")
	filename, err := d.sqliteLogger.Clone()
	if err != nil {
		fmt.Fprintf(w, "error: %s", err)
		return
	}
	fmt.Println("Done cloning and compressing the database")

	fn := filepath.Base(filename)
	file, err := os.Open(fn)
	if err != nil {
		fmt.Fprintf(w, "error: %s", err)
		return
	}
	_, err = io.Copy(w, file)
	if err != nil {
		fmt.Fprintf(w, "error: %s", err)
		return
	}
	file.Close()

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/x-gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fn))
	w.Write([]byte(fn))

	return
}
