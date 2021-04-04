.PHONY: all
all: test install image

.PHONY: install
install:
	CGO_ENABLED=0 GO111MODULE=on go build -a -ldflags="-X 'github.com/ssup2/kpexec/pkg/cmd/kpexec.version=latest'" -o ${GOBIN}/kpexec ./cmd/kpexec
	CGO_ENABLED=0 GO111MODULE=on go build -a -ldflags="-X 'github.com/ssup2/kpexec/pkg/cmd/kpexec.version=latest' -X 'github.com/ssup2/kpexec/pkg/cmd/kpexec.build=kubectlPlugin'" -o ${GOBIN}/kubectl-pexec ./cmd/kpexec

.PHONY: image
image:
	docker build --build-arg VERSION=latest -f Dockerfile-cnsenter -t ssup2/cnsenter:latest .
	docker build --build-arg VERSION=latest -f Dockerfile-cnsenter-tools -t ssup2/cnsenter-tools:latest .

.PHONY: clean
clean:
	rm -f kpexec

.PHONY: test
test:
	go test -v ./...
