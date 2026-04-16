VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/joeyfurness/rdl/cmd.version=$(VERSION) \
	-X github.com/joeyfurness/rdl/cmd.commit=$(COMMIT) \
	-X github.com/joeyfurness/rdl/cmd.date=$(DATE)

.PHONY: build test install clean

build:
	go build -ldflags "$(LDFLAGS)" -o rdl .

test:
	go test ./...

install:
	go install -ldflags "$(LDFLAGS)" .

clean:
	rm -f rdl
