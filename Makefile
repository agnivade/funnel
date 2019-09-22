export GO111MODULE=on

all: test install

build:
	go build ./cmd/funnel

install:
	go install ./cmd/funnel

lint:
	gofmt -l -s -w . && go vet -all ./... && golint -set_exit_status=1 ./...

test:
	go test -race -v -coverprofile=coverage.txt -covermode=atomic

bench:
	go test -run=XXX -bench=Processor -benchmem

release:
	GOOS=darwin GOARCH=amd64 go build -o funnel_darwin-amd64 -ldflags "-s -w" ./cmd/funnel
	GOOS=darwin GOARCH=amd64 go build -tags "disableelasticsearch disableinfluxdb disablekafka disableredis disables3 disablenats" -o funnel_minimal_darwin-amd64 -ldflags "-s -w" ./cmd/funnel
	GOOS=linux GOARCH=arm64 go build -o funnel_linux-arm64 -ldflags "-s -w" ./cmd/funnel
	GOOS=linux GOARCH=arm64 go build -tags "disableelasticsearch disableinfluxdb disablekafka disableredis disables3 disablenats" -o funnel_minimal_linux-arm64 -ldflags "-s -w" ./cmd/funnel
	GOOS=linux GOARCH=amd64 go build -o funnel_linux-amd64 -ldflags "-s -w" ./cmd/funnel
	GOOS=linux GOARCH=amd64 go build -tags "disableelasticsearch disableinfluxdb disablekafka disableredis disables3 disablenats" -o funnel_minimal_linux-amd64 -ldflags "-s -w" ./cmd/funnel
