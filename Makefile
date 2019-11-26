BINARY_NAME=mackerel-plugin-aws-transitgateway
PLATFORMS=linux darwin

.PHONY: init
init:
	GO111MODULE=on go mod init

.PHONY: deps
deps:
	GO111MODULE=off go get golang.org/x/lint/golint

.PHONY: lint
lint: deps
	go vet -all .
	golint -set_exit_status .

.PHONY: test
test:
	go test -v

.PHONY: build
build: dist/$(BINARY_NAME)_linux_amd64/$(BINARY_NAME) dist/$(BINARY_NAME)_darwin_amd64/$(BINARY_NAME)
dist/$(BINARY_NAME)_linux_amd64/$(BINARY_NAME): main.go $(shell find lib -type f)
	GOOS=linux GOARCH=amd64 go build -o $@ -v
dist/$(BINARY_NAME)_darwin_amd64/$(BINARY_NAME): main.go $(shell find lib -type f)
	GOOS=darwin GOARCH=amd64 go build -o $@ -v
