ARG GO_VERSION=1.24.2
ARG GIT_HASH
ARG VERSION_TAG

#------ Base
FROM golang:${GO_VERSION}-bookworm AS base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

#------ Build
FROM base AS go_builder

ARG BUILD_STAMP
ARG GIT_HASH
ARG VERSION_TAG

WORKDIR /app
COPY . /app/
RUN go build --trimpath -ldflags "\
      -X github.com/vkumov/go-pxgrider/server/internal/config.BuildStamp=${BUILD_STAMP} \
      -X github.com/vkumov/go-pxgrider/server/internal/config.GitHash=${GIT_HASH} \
      -X github.com/vkumov/go-pxgrider/server/internal/config.V=${VERSION_TAG} \
    " -buildvcs=false -o ./bin/pxgrider ./server/bin/*.go

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates

#------ Deploy
FROM debian:12-slim AS server
WORKDIR /app

COPY ./config.base.yml /app/config.yml
COPY --from=go_builder /app/bin/pxgrider /app/
COPY --from=go_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 50051

ENTRYPOINT ["/app/pxgrider"]