package logger

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSqlite_Purge(t *testing.T) {

	dbPath := "/tmp/test.db"
	_ = os.Remove(dbPath)

	db := NewSqlite(dbPath)
	err := db.Init(0)
	require.NoError(t, err)

	data := NewLoggerData()
	data.Timestamp = time.Now()

	err = db.Log(data)
	require.NoError(t, err)

	data.Timestamp = time.Now().Add(-time.Hour * 1)

	err = db.Log(data)
	require.NoError(t, err)

	res, err := db.db.Query(`SELECT count(*) as c FROM gnss`)
	require.NoError(t, err)
	count := 0
	n := res.Next()
	require.True(t, n)
	err = res.Scan(&count)
	res.Close()
	require.NoError(t, err)
	require.Equal(t, 2, count)

	err = db.Purge(30 * time.Minute)
	require.NoError(t, err)

	res, err = db.db.Query("SELECT * FROM gnss")
	require.NoError(t, err)

	res, err = db.db.Query(`SELECT count(*) as c FROM gnss`)
	require.NoError(t, err)
	count = 0
	n = res.Next()
	require.True(t, n)
	err = res.Scan(&count)
	res.Close()
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestSqlite_7deg(t *testing.T) {
	f := 45.7648778
	i := int32(f * 1e7)
	require.Equal(t, int32(457648778), i)
}
