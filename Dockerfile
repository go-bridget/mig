FROM golang:1.14-alpine AS build

ADD . /root/mig

WORKDIR /root/mig

RUN CGO_ENABLED=0 go build ./cmd/...

FROM alpine:latest

COPY --from=build /root/mig/mig /usr/local/bin/mig

WORKDIR /app

ENTRYPOINT ["mig"]
