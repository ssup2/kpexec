.PHONY: test build image
all: test build image

build:
	GO111MODULE=on CGO_ENABLED=0 go build -a -mod vendor -o kpexec ./cmd/kpexec/main.go
	GO111MODULE=on CGO_ENABLED=0 go build -a -mod vendor -o cnsenter ./cmd/cnsenter/main.go

image:
	docker build -f build/nodepause/Dockerfile -t ssup2/nodepause:latest .

clean:
	rm -f kpexec cnsenter

test:
	go test -v ./...
