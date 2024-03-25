########
# BASE #
########
ARG NODE_VERSION="20"
ARG GO_VERSION="1.22"
FROM golang:${GO_VERSION}-alpine AS go_builder
FROM node:${NODE_VERSION}-alpine AS base

COPY --from=go_builder /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"
RUN npm install -g npm@latest

RUN \
  apk update && \
  apk add --no-cache \
    make

COPY . /build/
WORKDIR /build/

ARG BUILD_VERSION="development"
RUN make BUILD_VERSION=${BUILD_VERSION} build
RUN chmod 755 argus
RUN chmod 755 healthcheck


#########
# ARGUS #
#########
FROM alpine:latest
LABEL maintainer="The Argus Authors <developers@release-argus.io>"
RUN \
  apk update && \
  apk add --no-cache \
    ca-certificates \
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
