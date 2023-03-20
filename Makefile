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

docker: export GOOS = linux
docker: export GOARCH = amd64
docker: build
	docker build -t jaconi.io/secret-file-provider:latest .

mod:
	go mod tidy
