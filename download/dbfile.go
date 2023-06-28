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

func (d *Download) GetDatabaseFiles(w http.ResponseWriter, _ *http.Request) {
	filename, err := d.sqliteLogger.Clone()
	if err != nil {
		fmt.Fprintf(w, "error: %s", err)
		return
	}

	fn := filepath.Base(filename)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fn))
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

	fmt.Fprintf(w, "Successfully Downloaded %s", filename)
}
