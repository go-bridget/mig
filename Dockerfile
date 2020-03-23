FROM golang:1.14-alpine AS build

ADD . /root/mig

ARG GIT_COMMIT

WORKDIR /root/mig

RUN apk --no-cache add git
RUN export GIT_COMMIT=$(git rev-list -1 HEAD)
RUN CGO_ENABLED=0 go build -ldflags "-X main.Version=$GIT_COMMIT" ./cmd/...

FROM alpine:latest

COPY --from=build /root/mig/mig /usr/local/bin/mig

WORKDIR /app

ENTRYPOINT ["mig"]
