module github.com/streamingfast/hivemapper-data-logger

go 1.20

replace github.com/daedaleanai/ublox => github.com/streamingfast/ublox v0.0.0-20230531132611-c1a91eeb44cf

replace github.com/streamingfast/imu-controller => ../imu-controller

replace github.com/streamingfast/gnss-controller => ../gnss-controller

require (
	github.com/charmbracelet/bubbletea v0.23.1
	github.com/charmbracelet/lipgloss v0.7.1
	github.com/spf13/cobra v1.7.0
	github.com/streamingfast/gnss-controller v0.1.20-0.20230531213918-8008ed1d7e97
	github.com/streamingfast/imu-controller v0.0.0-20230531144200-f741eb6d9f3d
	github.com/stretchr/testify v1.8.4
	modernc.org/sqlite v1.22.1
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/containerd/console v1.0.3 // indirect
	github.com/daedaleanai/ublox v0.0.0-00010101000000-000000000000 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-localereader v0.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/muesli/ansi v0.0.0-20211018074035-2e021307bc4b // indirect
	github.com/muesli/cancelreader v0.2.2 // indirect
	github.com/muesli/reflow v0.3.0 // indirect
	github.com/muesli/termenv v0.15.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/tarm/serial v0.0.0-20180830185346-98f6abe2eb07 // indirect
	golang.org/x/mod v0.3.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.0.0-20201124115921-2c860bdd6e78 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.40.0 // indirect
	modernc.org/ccgo/v3 v3.16.13 // indirect
	modernc.org/libc v1.22.5 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.0.1 // indirect
	periph.io/x/conn/v3 v3.7.0 // indirect
	periph.io/x/host/v3 v3.8.2 // indirect
)
