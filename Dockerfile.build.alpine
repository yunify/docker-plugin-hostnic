FROM golang:1.7.3-alpine

ENV GOPATH /go

RUN mkdir -p "$GOPATH/src/" "$GOPATH/bin" && chmod -R 777 "$GOPATH" && \
    mkdir -p /go/src/github.com/yunify/docker-plugin-hostnic

RUN apk --update add bash git gcc
RUN apk add --update alpine-sdk
RUN apk add --update linux-headers
RUN ln -s /go/src/github.com/yunify/docker-plugin-hostnic /app

WORKDIR /app
