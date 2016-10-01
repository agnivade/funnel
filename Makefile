all: test install

build:
	go build ./cmd/funnel

install:
	go install ./cmd/funnel

test:
	gofmt -l -s -w . && go tool vet -all . && go test -race -v