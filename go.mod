module github.com/Hivemapper/hivemapper-data-logger

go 1.20

//replace github.com/daedaleanai/ublox => /Users/cbillett/devel/github/ublox
// replace github.com/daedaleanai/ublox => github.com/streamingfast/ublox v0.0.0-20230815154721-b29363712a91
replace github.com/daedaleanai/ublox => github.com/Hivemapper/sf-ublox v0.0.0-20240221201612-d92d22b86230

replace github.com/Hivemapper/gnss-controller => ../gnss-controller

require (
	github.com/Hivemapper/gnss-controller v1.0.3-0.20240402232423-1de9f3a3a7f8
	github.com/google/uuid v1.3.1
	github.com/rosshemsley/kalman v0.0.0-20190615074247-f4b900823fd1
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.4
	modernc.org/sqlite v1.26.0
	periph.io/x/conn/v3 v3.7.0
	periph.io/x/host/v3 v3.8.2
)

require (
	github.com/daedaleanai/ublox v0.0.0-20210116232802-16609b0f9f43 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/streamingfast/shutter v1.5.0 // indirect
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07 // indirect
	golang.org/x/exp v0.0.0-20240325151524-a685a6edb6d8
	golang.org/x/mod v0.16.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
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
)
