all: test install

build:
	go build ./cmd/funnel

install:
	go install ./cmd/funnel

lint:
	gofmt -l -s -w . && go tool vet -all . && golint

test:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic

release:
	GOOS=darwin GOARCH=amd64 go build -o funnel_darwin-amd64 ./cmd/funnel
	GOOS=linux GOARCH=arm64 go build -o funnel_linux-arm64 ./cmd/funnel
	GOOS=linux GOARCH=amd64 go build -o funnel_linux-amd64 ./cmd/funnel