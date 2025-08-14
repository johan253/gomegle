FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache make

COPY go.mod go.sum Makefile ./

RUN make deps

COPY . .

RUN make build

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates ncurses ncurses-terminfo-base ncurses-terminfo

COPY --from=builder /app/bin/ /app/bin/

EXPOSE 23234

ENTRYPOINT ["/app/bin/gomegle"]
