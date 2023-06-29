package download

import (
	"fmt"
	"github.com/streamingfast/hivemapper-data-logger/logger"
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

func (d *Download) GetDatabaseFiles(w http.ResponseWriter, r *http.Request) {
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
