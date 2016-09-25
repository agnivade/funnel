all: build

build:
	go build ./cmd/funnel

install:
	go install ./cmd/funnel

test:
	go test