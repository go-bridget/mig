FROM alpine:latest

ADD ./build/mig /usr/local/bin/mig
RUN mig version

WORKDIR /app
ENTRYPOINT ["mig"]
