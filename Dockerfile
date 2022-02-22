FROM golang:1.17-alpine AS build

ADD . /root/mig
WORKDIR /root/mig
RUN apk --no-cache add git make
RUN make build

FROM alpine:latest as test

COPY --from=build /root/mig/_build/mig /usr/local/bin/mig
RUN mig version

FROM alpine:latest

COPY --from=build /root/mig/_build/mig /usr/local/bin/mig
WORKDIR /app
ENTRYPOINT ["mig"]
