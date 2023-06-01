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
