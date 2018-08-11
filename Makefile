all: test install

build:
	go build ./cmd/funnel

get-dep:
	go get github.com/fsnotify/fsnotify
	go get github.com/spf13/viper
	go get vbom.ml/util/sortorder
	go get golang.org/x/net/context
	go get github.com/Shopify/sarama
	go get gopkg.in/olivere/elastic.v5
	go get github.com/influxdata/influxdb/client/v2
	go get gopkg.in/redis.v5
	go get github.com/aws/aws-sdk-go
	go get github.com/nats-io/go-nats

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
