module github.com/Hivemapper/gnss-controller

go 1.20

//replace github.com/daedaleanai/ublox => /Users/cbillett/devel/github/ublox
replace github.com/daedaleanai/ublox => github.com/streamingfast/ublox v0.0.0-20230815154721-b29363712a91

require (
	github.com/daedaleanai/ublox v0.0.0-00010101000000-000000000000
	github.com/streamingfast/shutter v1.5.0
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07
)

require golang.org/x/sys v0.0.0-20220811171246-fbc7d0a398ab // indirect
