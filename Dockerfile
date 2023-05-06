############################
# STEP 1 build executable binary
############################
FROM golang:1.20.3-alpine3.17 AS builder

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

ENV GOPATH=/projects/go

WORKDIR $GOPATH/src/github.com/knave-de-coeur/wallet-api/

COPY . .
# Fetch dependencies.
# Using go get.
RUN go mod tidy
# Build the binary.
RUN go build -o $GOPATH/src/github.com/knave-de-coeur/wallet-api/bin/wallet-api $GOPATH/src/github.com/knave-de-coeur/wallet-api/cmd/api/main.go
############################
# STEP 2 build a small image
############################
FROM scratch

ENV GOPATH=/projects/go

# Copy our static executable.
COPY --from=builder $GOPATH/src/github.com/knave-de-coeur/wallet-api/bin/wallet-api $GOPATH/src/github.com/knave-de-coeur/wallet-api/bin/wallet-api
COPY --from=builder $GOPATH/src/github.com/knave-de-coeur/wallet-api/internal/migrations $GOPATH/src/github.com/knave-de-coeur/wallet-api/bin/migrations

EXPOSE 8080

# Run the api binary.
ENTRYPOINT ["/projects/go/src/github.com/knave-de-coeur/wallet-api/bin/wallet-api"]
