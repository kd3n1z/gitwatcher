all: build

build: clean
	-mkdir dest
	go build -o "dest/gitwatcher"

build-all:
	-mkdir dest
	-mkdir dest/linux
	-mkdir dest/linux/arm64
	-mkdir dest/linux/x64
	env GOOS=linux GOARCG=arm64 go build -o "dest/linux/arm64/gitwatcher"
	env GOOS=linux GOARCG=amd64 go build -o "dest/linux/x64/gitwatcher"
	-mkdir dest/darwin
	-mkdir dest/darwin/x64
	env GOOS=darwin GOARCG=amd64 go build -o "dest/darwin/x64/gitwatcher"

alias:
	echo alias gwgo="/Users/deniskomarkov/Documents/gitwatcher-go/dest/gitwatcher"

clean: 
	-rm -rf dest