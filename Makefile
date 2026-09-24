
OUT := stew

.PHONY: build test integration smoke vet lint

VERSION ?= dev
LDFLAGS := -X github.com/benjuh/stew/cmd.Version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o $(OUT) .

test:
	go test ./...

integration:
	go test ./integration -run TestCLIWorkflow -v

smoke: integration

vet:
	go vet ./...

lint:
	test -z "$(gofmt -l .)"
