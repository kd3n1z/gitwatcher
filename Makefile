zipArgs=-9
goArgs=-ldflags="-s -w -X main.COMMIT=$$(git rev-parse HEAD) -X main.BRANCH=$$(git rev-parse --abbrev-ref HEAD)" ./src/gitwatcher

all: build

build: clean
	-mkdir dest
	go build -o "dest/gitwatcher" $(goArgs)

build-all: clean
	-mkdir dest
	-mkdir dest/linux
	-mkdir dest/linux/arm64
	-mkdir dest/linux/x64
	-mkdir dest/darwin
	-mkdir dest/darwin/x64
	-mkdir dest/windows
	-mkdir dest/windows/arm64
	-mkdir dest/windows/x64
	env GOOS=linux GOARCH=arm64 go build -o "dest/linux/arm64/gitwatcher" $(goArgs)
	env GOOS=linux GOARCH=amd64 go build -o "dest/linux/x64/gitwatcher" $(goArgs)
	env GOOS=darwin GOARCH=amd64 go build -o "dest/darwin/x64/gitwatcher" $(goArgs)
	env GOOS=windows GOARCH=amd64 go build -o "dest/windows/x64/gitwatcher.exe" $(goArgs)
	env GOOS=windows GOARCH=arm64 go build -o "dest/windows/arm64/gitwatcher.exe" $(goArgs)

zip-all:
	-mkdir dest/zips
	cd dest/linux/arm64; zip $(zipArgs) gitwatcher gitwatcher; cp gitwatcher.zip ../../zips/linux-arm64.zip
	cd dest/linux/x64; zip $(zipArgs) gitwatcher gitwatcher; cp gitwatcher.zip ../../zips/linux-x64.zip
	cd dest/windows/arm64; zip $(zipArgs) gitwatcher gitwatcher.exe; cp gitwatcher.zip ../../zips/windows-arm64.zip
	cd dest/windows/x64; zip $(zipArgs) gitwatcher gitwatcher.exe; cp gitwatcher.zip ../../zips/windows-x64.zip
	cd dest/darwin/x64; zip $(zipArgs) gitwatcher gitwatcher; cp gitwatcher.zip ../../zips/macOS.zip

publish-all: build-all zip-all

clean: 
	-rm -rf dest
