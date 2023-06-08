# hivemapper data logger

## Install buf
Make sure you have buf installed
```bash
brew install bufbuild/buf/buf
```

## Install protoc-gen
Make sure you have installed `protoc-gen-go` and `protoc-gen-connect-go`
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest
```

## Generate proto
Once you have installed `protoc-gen-go` and `protoc-gen-connect-go`, generate the proto files with:
```bash
make generate
```

## Build datalogger
This project is used for the dashcams, so you need to build for a specific architecture
```bash 
GOOS=linux GOARCH=arm64 go build ./cmd/datalogger
```

## Run on the camera
```bash
./datalogger wip --db-output-path=/path/to/output.db
# db-output-path is the location to where we want the imu and gnss events to be saved
```

### Run the replay command with events which were saved to a sqlite file
Once you have run the command above to run on the camera, all the events that you have emitted, they will be saved to a sqlite database. Given the path of where the sqlite has saved the events, then we can rerun the _car run_ instead of going back out and driving. Permits to easily iterate on data.
```bash
datalogger replay --clean --db-import-path=/path/to/databse --db-output-path=/tmp/out.db
# clean will delete the db-output-path prior to running the replay command -> this is good to remove previous runs
# db-import-path will take the .db file which was produced by the datalogger log command
# db-output-path is the location to where the rerun db will be saved
```