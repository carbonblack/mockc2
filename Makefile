VERSION=$(shell git describe --tags)
BUILD_DATE=`date +%FT%T%z`
GOLDFLAGS += -s -w
GOLDFLAGS += -X megaman.genesis.local/sknight/mockc2/pkg/version.Version=$(VERSION)
GOLDFLAGS += -X megaman.genesis.local/sknight/mockc2/pkg/version.BuildDate=$(BUILD_DATE)
GOFLAGS = -ldflags "$(GOLDFLAGS)"

darwin64:
	cd cmd/mockc2 && GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -o ../../build/darwin-amd64/mockc2

linux64:
	cd cmd/mockc2 && GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -o ../../build/linux-amd64/mockc2

windows64:
	cd cmd/mockc2 && GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -o ../../build/windows-amd64/mockc2.exe

clean:
	rm -rf build

all: darwin64 linux64 windows64
