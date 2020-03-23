.PHONY: all docker push test

IMAGE := bridget/mig

all: docker test

docker:
	docker build --no-cache --build-arg GIT_COMMIT=$(shell git rev-list -1 HEAD) --rm -t $(IMAGE) .

push:
	docker push $(IMAGE)

test:
	drone exec --trusted