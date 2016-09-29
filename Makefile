all: test install

build:
	go build ./cmd/funnel

install:
	go install ./cmd/funnel

test:
	go test -race -v