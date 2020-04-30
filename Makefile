.PHONY: all docker push build test

IMAGE := bridget/mig

all: docker test

export BUILD_VERSION := $(shell git describe --always --tags --abbrev=8)
export BUILD_TIME := $(shell date +%Y-%m-%dT%T%z)

docker:
	docker build --no-cache --rm -t $(IMAGE) .

push:
	docker push $(IMAGE)

build:
	CGO_ENABLED=0 go build -ldflags "-X 'main.BuildVersion=$(BUILD_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'" ./cmd/...

test:
	drone exec --trusted