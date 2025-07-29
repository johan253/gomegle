FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum Makefile ./

RUN apk add --no-cache make && make deps

COPY . .

RUN make build

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/ /app/bin/

ENTRYPOINT ["/app/bin/gomegle"]
