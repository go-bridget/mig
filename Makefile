.PHONY: all docker push build test

IMAGE := gobridget/mig

all: docker push test

export BUILD_VERSION := $(shell git describe --always --tags --abbrev=8)
export BUILD_TIME := $(shell date +%Y-%m-%dT%T%z)
export CGO_ENABLED := 0

docker:
	docker build --no-cache --rm -t $(IMAGE) .

push:
	docker push $(IMAGE)

build:
	go fmt ./...
	mkdir -p build
	go build -o build -ldflags "-X 'main.BuildVersion=$(BUILD_VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'" ./cmd/...

test:
	drone exec --trusted