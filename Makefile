.PHONY: all
all: test install image

.PHONY: install
install:
	CGO_ENABLED=0 GO111MODULE=on go install -a -ldflags="-X 'github.com/ssup2/kpexec/pkg/cmd/kpexec.version=latest'" ./cmd/kpexec

.PHONY: image
image:
	docker build -f Dockerfile-cnsenter -t ssup2/cnsenter:latest .
	docker build -f Dockerfile-cnsenter-tools -t ssup2/cnsenter-tools:latest .

.PHONY: clean
clean:
	rm -f kpexec

.PHONY: test
test:
	go test -v ./...
