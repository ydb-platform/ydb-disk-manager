FROM golang:1.24-bookworm as build

WORKDIR /go/app
COPY . /go/app

RUN apt update && apt install -y libudev-dev
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o ydb-disk-manager cmd/ydb-disk-manager/main.go

FROM debian:bookworm

WORKDIR /root

RUN apt update && apt install -y libudev1 && apt clean
COPY --from=build /go/app/ydb-disk-manager /usr/bin/ydb-disk-manager

ENTRYPOINT ["/usr/bin/ydb-disk-manager"]
