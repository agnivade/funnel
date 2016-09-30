all: test install

build:
	go build ./cmd/funnel

install:
	go install ./cmd/funnel

test:
	go vet && go test -race -v