module github.com/Hivemapper/hivemapper-data-logger

go 1.20

//replace github.com/daedaleanai/ublox => /Users/cbillett/devel/github/ublox
// replace github.com/daedaleanai/ublox => github.com/streamingfast/ublox v0.0.0-20230815154721-b29363712a91
replace github.com/daedaleanai/ublox => ./sf-ublox

// hack: find a better way to do this
replace github.com/Hivemapper/gnss-controller => ./gnss-controller

replace github.com/streamingfast/imu-controller => ./imu-controller

require (
	github.com/Hivemapper/gnss-controller v1.0.3-0.20240819070221-78cf51b8a5c6
	github.com/bufbuild/connect-go v1.10.0
	github.com/fsnotify/fsnotify v1.6.0
	github.com/google/uuid v1.3.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/paulmach/go.geojson v1.5.0
	github.com/rosshemsley/kalman v0.0.0-20190615074247-f4b900823fd1
	github.com/rs/cors v1.10.1
	github.com/spf13/cobra v1.7.0
	github.com/streamingfast/imu-controller v0.0.0-20230928133410-7c6595dd3783
	github.com/stretchr/testify v1.8.4
	go.mongodb.org/mongo-driver v1.12.1
	google.golang.org/protobuf v1.31.0
	modernc.org/sqlite v1.26.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
)

require (
	github.com/daedaleanai/ublox v0.0.0-20210116232802-16609b0f9f43 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.17.1 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/streamingfast/shutter v1.5.0 // indirect
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20201027041543-1326539a0a0a // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/exp v0.0.0-20240325151524-a685a6edb6d8
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/tools v0.19.0 // indirect
	gonum.org/v1/gonum v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/uint128 v1.3.0 // indirect
	modernc.org/cc/v3 v3.41.0 // indirect
	modernc.org/ccgo/v3 v3.16.15 // indirect
	modernc.org/libc v1.24.1 // indirect
	modernc.org/mathutil v1.6.0 // indirect
	modernc.org/memory v1.7.2 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/strutil v1.2.0 // indirect
	modernc.org/token v1.1.0 // indirect
	periph.io/x/conn/v3 v3.7.0 // indirect
	periph.io/x/host/v3 v3.8.2 // indirect
)
