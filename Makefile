BINARY=raspiweather
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-w -s -X main.Version=$(VERSION)"

all: build

build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY)

build-linux-amd64:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY)-linux-amd64

build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY)-linux-arm64

build-linux-arm7:
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY)-linux-arm7

build-linux-arm6:
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY)-linux-arm6

build-linux-arm5:
	GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY)-linux-arm5

build-all: build-linux-amd64 build-linux-arm64 build-linux-arm7 build-linux-arm6 build-linux-arm5

clean:
	rm -f $(BINARY) $(BINARY)-*

.PHONY: all build build-linux-amd64 build-linux-arm64 build-linux-arm7 build-linux-arm6 build-linux-arm5 build-all clean