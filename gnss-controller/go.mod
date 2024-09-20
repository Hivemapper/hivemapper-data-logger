module github.com/Hivemapper/gnss-controller

go 1.20

//replace github.com/daedaleanai/ublox => /Users/cbillett/devel/github/ublox
// replace github.com/daedaleanai/ublox => github.com/Hivemapper/sf-ublox/tree/main
// replace github.com/daedaleanai/ublox => github.com/streamingfast/ublox v0.0.0-20230815154721-b29363712a91
replace github.com/daedaleanai/ublox => github.com/Hivemapper/sf-ublox v0.0.0-20240221201612-d92d22b86230

require (
	github.com/daedaleanai/ublox v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.7.0
	github.com/streamingfast/shutter v1.5.0
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
)

