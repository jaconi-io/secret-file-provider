# Image URL to use all building/pushing image targets
IMG ?= jaconi.io/secret-file-provider:latest

all: test

run:
	go run main.go

test: build
	go test . ./pkg/... -coverprofile cover.out

build: fmt vet
	rm -f bin/operator
	go build -o bin/operator main.go

fmt: mod
	go fmt . ./pkg/...

vet:
	go vet ./pkg/...

docker-build: build
	docker build -t ${IMG} .

docker-push:
	docker push ${IMG}

mod:
	go mod tidy
