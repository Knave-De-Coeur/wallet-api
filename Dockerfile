FROM golang:1.20.3-alpine3.17 AS build-env

COPY . .

WORKDIR /go/src/github.com/knave-de-coeur/wallet-api

# Download necessary Go modules
RUN go mod tidy
