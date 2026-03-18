FROM golang:1.26-alpine AS builder

ENV TERM=linux

WORKDIR /go/src/goprometrics

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o goprometrics .

FROM alpine:3.23

COPY --from=builder /go/src/goprometrics/goprometrics /usr/local/bin/goprometrics

EXPOSE 9111 9112

ENTRYPOINT ["goprometrics"]
