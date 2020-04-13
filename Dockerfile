FROM golang:1.14.2-alpine

ENV TERM=linux

RUN apk update && \
    apk upgrade && \
    apk add git

COPY ./src /go/src/jtlprom
WORKDIR  /go/src/jtlprom

RUN go get ./
RUN go build
RUN go get github.com/pilu/fresh

ENTRYPOINT fresh

EXPOSE 9111 9112