package logger

//func TestSqlite_Purge(t *testing.T) {
//
//	dbPath := "/tmp/test.sqliteLogger"
//	_ = os.Remove(dbPath)
//
//	sqliteLogger := NewSqlite(dbPath)
//	err := sqliteLogger.Init(0, &data.Subscription{})
//	require.NoError(t, err)
//
//	sqliteLogger.StartStoring()
//
//	df := neom9n.NewDataFeed(nil)
//	df.Data.Timestamp = time.Now()
//
//	err = sqliteLogger.Log(df.Data)
//	require.NoError(t, err)
//
//	df.Data.Timestamp = time.Now().Add(-time.Hour * 1)
//
//	err = sqliteLogger.Log(df.Data)
//	require.NoError(t, err)
//
//	res, err := sqliteLogger.DB.Query(`SELECT count(*) as c FROM gnss`)
//	require.NoError(t, err)
//	count := 0
//	n := res.Next()
//	require.True(t, n)
//	err = res.Scan(&count)
//	res.Close()
//	require.NoError(t, err)
//	require.Equal(t, 2, count)
//
//	err = sqliteLogger.Purge(30 * time.Minute)
//	require.NoError(t, err)
//
//	res, err = sqliteLogger.DB.Query("SELECT * FROM gnss")
//	require.NoError(t, err)
//
//	res, err = sqliteLogger.DB.Query(`SELECT count(*) as c FROM gnss`)
//	require.NoError(t, err)
//	count = 0
//	n = res.Next()
//	require.True(t, n)
//	err = res.Scan(&count)
//	res.Close()
//	require.NoError(t, err)
//	require.Equal(t, 1, count)
//}
//
//func TestSqlite_7deg(t *testing.T) {
//	f := 45.7648778
//	i := int32(f * 1e7)
//	require.Equal(t, int32(457648778), i)
//}
