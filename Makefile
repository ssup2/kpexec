.PHONY: test build image
all: test build image

install:
	CGO_ENABLED=0 GO111MODULE=on go install -a ./cmd/kpexec

image:
	docker build -f Dockerfile-cnsenter -t ssup2/cnsenter:latest .
	docker build -f Dockerfile-cnsenter-tools -t ssup2/cnsenter-tools:latest .

clean:
	rm -f kpexec

test:
	go test -v ./...
