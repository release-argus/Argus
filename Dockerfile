########
# BASE #
########
ARG DEBIAN_VERSION="bookworm"
ARG GO_VERSION="1.25"
ARG NODE_VERSION="24"
ARG ALPINE_VERSION="3.23@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375"
FROM golang:${GO_VERSION}-${DEBIAN_VERSION} AS go_builder
FROM node:${NODE_VERSION}-${DEBIAN_VERSION} AS base

COPY --from=go_builder /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"
RUN npm install -g npm@latest

COPY . /build/
WORKDIR /build/

ARG BUILD_VERSION="development"
RUN make BUILD_VERSION=${BUILD_VERSION} build \
  && chmod 755 /build/argus \
  && chmod 755 /build/healthcheck


#########
# ARGUS #
#########
FROM alpine:${ALPINE_VERSION}
LABEL maintainer="The Argus Authors <developers@release-argus.io>"
RUN \
  apk update && \
  apk add --no-cache \
    ca-certificates \
    curl \
    musl-dev \
    su-exec && \
  rm -rf \
    /tmp/* \
    /var/cache/*

COPY entrypoint.sh /entrypoint.sh
COPY --from=base /build/argus              /usr/bin/argus
COPY --from=base /build/healthcheck        /healthcheck
COPY --from=base /build/config.yml.example /app/config.yml
COPY --from=base /build/LICENSE            /LICENSE

RUN \
  addgroup -g 911 -S argus && \
  adduser -u 911 -S -D -h /app -s /bin/false argus -G argus && \
  mkdir -p \
    /app \
    /app/data && \
  chown -R argus:argus /app
WORKDIR /app

EXPOSE     8080
VOLUME     [ "/app/data" ]
ENTRYPOINT [ "/entrypoint.sh" ]
