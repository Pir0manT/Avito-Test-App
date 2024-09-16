BINARY_NAME=myapp

all: windows linux mac

windows:
    GOOS=windows GOARCH=amd64 go build -o bin/windows/$(BINARY_NAME).exe

linux:
    GOOS=linux GOARCH=amd64 go build -o bin/linux/$(BINARY_NAME)

mac:
    GOOS=darwin GOARCH=amd64 go build -o bin/mac/$(BINARY_NAME)