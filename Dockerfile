# Builder
FROM golang:1.23-alpine3.21 as builder

ARG VERSION="0.0.0-build"
ENV VERSION=$VERSION

WORKDIR /go/src/app

RUN apk add --no-cache \
  bash build-base git make

# Copy module files and download dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Then copy the rest of the source code and build
COPY . ./
RUN make build

# Final image
FROM alpine:3.21
RUN apk upgrade --no-cache \
  && apk add --no-cache bash curl \
  && apk add --no-cache --upgrade libssl3 libcrypto3

RUN addgroup -g 1001 app \
  && adduser -D -G app -u 1001 app

COPY --from=builder /go/src/app/build/cosmos-validator-watcher /
USER app

WORKDIR /
ENTRYPOINT ["/cosmos-validator-watcher"]
