# Release

## Build the datalogger and make sure it runs as expected
```bash
GOOS=linux GOARCH=arm64 go build ./cmd/datalogger
```
Then run the datalogger on the camera with
```bash
./datalogger log
```

## Create git tag and push
Make sure you are not on a dirty state of git and that you have pushed everything
```bash
git tag <version>
git push origin --tags 
```

## Run goreleaser
```bash
rm -rf dist
goreleaser release
```
