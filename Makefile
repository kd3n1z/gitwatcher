zipArgs=-9
goArgs=-ldflags="-s -w -X main.COMMIT=$$(git rev-parse HEAD)"

all: build

build: clean
	-mkdir dest
	go build $(goArgs) -o "dest/gitwatcher"

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
	env GOOS=linux GOARCG=arm64 go build $(goArgs) -o "dest/linux/arm64/gitwatcher"
	env GOOS=linux GOARCG=amd64 go build $(goArgs) -o "dest/linux/x64/gitwatcher"
	env GOOS=darwin GOARCG=amd64 go build $(goArgs) -o "dest/darwin/x64/gitwatcher"
	env GOOS=windows GOARCG=amd64 go build $(goArgs) -o "dest/windows/x64/gitwatcher.exe"
	env GOOS=windows GOARCG=arm64 go build $(goArgs) -o "dest/windows/arm64/gitwatcher.exe"

zip-all:
	-mkdir dest/zips
	cd dest/linux/arm64; zip $(zipArgs) gitwatcher gitwatcher; cp gitwatcher.zip ../../zips/linux-arm64.zip
	cd dest/linux/x64; zip $(zipArgs) gitwatcher gitwatcher; cp gitwatcher.zip ../../zips/linux-x64.zip
	cd dest/windows/arm64; zip $(zipArgs) gitwatcher gitwatcher.exe; cp gitwatcher.zip ../../zips/windows-arm64.zip
	cd dest/windows/x64; zip $(zipArgs) gitwatcher gitwatcher.exe; cp gitwatcher.zip ../../zips/windows-x64.zip
	cd dest/darwin/x64; zip $(zipArgs) gitwatcher gitwatcher; cp gitwatcher.zip ../../zips/macOS.zip

publish-all: build-all zip-all


alias:
	echo alias gwgo="/Users/deniskomarkov/Documents/gitwatcher-go/dest/gitwatcher"

clean: 
	-rm -rf dest