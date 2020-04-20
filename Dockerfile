FROM golang:1.14.2-alpine

ENV TERM=linux

COPY . /go/src/jtlprom
WORKDIR  /go/src/jtlprom

RUN go get ./
RUN go build

ENTRYPOINT ./goprometrics

EXPOSE 9111 9112