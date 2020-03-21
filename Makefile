.PHONY: all docker push test

IMAGE := bridget/mig

all: docker test

docker:
	docker build --no-cache --rm -t $(IMAGE) .

push:
	docker push $(IMAGE)

test:
	drone exec --trusted