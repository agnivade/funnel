all: test install

build:
	go build ./cmd/funnel

install:
	go install ./cmd/funnel

lint:
	gofmt -l -s -w . && go tool vet -all . && golint

test:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic
