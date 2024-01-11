image_name = cr.yandex/crpl7ipeu79oseqhcgn2/ydb-disk-manager
image_version = $(shell grep appVersion helm/ydb-disk-manager/Chart.yaml | cut -d ' ' -f 2)
charts_registry = cr.yandex/crpl7ipeu79oseqhcgn2/charts
helm_version = $(shell grep version helm/ydb-disk-manager/Chart.yaml | cut -d ' ' -f 2)
image = $(image_name):$(image_version)

SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build push clean

##@ General

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

##@ Build

.PHONY: build
build: fmt vet
	docker build -t $(image) -f Dockerfile .
	helm package helm/ydb-disk-manager

##@ Deployment

.PHONY: push
push:
	docker push $(image)
	helm push ydb-disk-manager-$(helm_version).tgz oci://$(charts_registry)/

.PHONY: clean
clean:
	rm ydb-disk-manager-$(helm_version).tgz
	docker rmi $(image)
