########
# BASE #
########
ARG DEBIAN_VERSION="bullseye"
ARG GO_VERSION="1.21"
FROM golang:${GO_VERSION}-${DEBIAN_VERSION}

COPY . /build/
WORKDIR /build/

ARG BUILD_VERSION="development"
RUN make BUILD_VERSION=${BUILD_VERSION} go-build
RUN chmod 755 /build/argus

WORKDIR /build/_healthcheck/
RUN go build -o ../healthcheck
RUN chmod 755 /build/healthcheck


#########
# ARGUS #
#########
ARG DEBIAN_VERSION="bullseye"
FROM debian:${DEBIAN_VERSION}-slim
LABEL maintainer="The Argus Authors <developers@release-argus.io>"
RUN \
    apt update && \
    apt install ca-certificates -y && \
    apt autoremove -y && \
    apt clean && \
    rm -rf \
      /tmp/* \
      /usr/share/doc/* \
      /var/cache/* \
      /var/lib/apt/lists/* \
      /var/tmp/*

COPY entrypoint.sh /entrypoint.sh
COPY --from=0 /build/argus               /usr/bin/argus
COPY --from=0 /build/healthcheck         /healthcheck
COPY --from=0 /build/config.yml.example  /app/config.yml
COPY --from=0 /build/LICENSE             /LICENSE

RUN \
    useradd -u 911 -U -d /app -s /bin/false argus && \
    mkdir -p \
        /app \
        /app/data
WORKDIR /app

EXPOSE     8080
VOLUME     [ "/app/data" ]
ENTRYPOINT [ "/entrypoint.sh" ]
