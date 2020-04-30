FROM golang:1.14-alpine AS build

ADD . /root/mig

WORKDIR /root/mig

RUN apk --no-cache add git make
RUN make build

FROM alpine:latest

COPY --from=build /root/mig/mig /usr/local/bin/mig

WORKDIR /app

ENTRYPOINT ["mig"]
