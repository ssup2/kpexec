FROM gcc:9.2 as gcc-builder
WORKDIR /app
COPY . .
RUN gcc -Os -Wall -Werror -static -o /app/pause /app/build/nodepause/pause.c

FROM golang:1.13.6 as golang-builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -mod vendor -o /app/cnsenter /app/cmd/cnsenter

FROM alpine:3.11.3
RUN apk --no-cache add ca-certificates && apk --no-cache add util-linux
COPY --from=gcc-builder /app/pause /pause
COPY --from=golang-builder /app/cnsenter /usr/bin/cnsenter
ENTRYPOINT ["/pause"]
