FROM golang:1.18 as build

WORKDIR /go/app
COPY . /go/app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ydb-disk-manager cmd/ydb-disk-manager/main.go

FROM ubuntu:20.04

WORKDIR /root

COPY --from=build /go/app/ydb-disk-manager /usr/bin/ydb-disk-manager

ENTRYPOINT ["/usr/bin/ydb-disk-manager"]
