########
# BASE #
########
ARG GO_VERSION="1.18.1"
ARG DEBIAN_VERSION="bullseye"
FROM golang:${GO_VERSION}-${DEBIAN_VERSION}
# Install node for Makefille.common's:
# $(shell node -p "require('./web/ui/react-app/package.json').$(1)")
RUN \
    apt-get update && \
    apt-get install nodejs -y

COPY . /build/
WORKDIR /build/

ARG BUILD_VERSION="development"
RUN make BUILD_VERSION=${BUILD_VERSION} go-build
RUN chmod 755 /build/hymenaios


#############
# HYMENAIOS #
#############
ARG DEBIAN_VERSION="bullseye"
FROM debian:${DEBIAN_VERSION}-slim
LABEL maintainer="The Hymenaios Authors <developers@hymenaios.io>"
RUN \
    apt-get update && \
    apt-get install ca-certificates -y && \
    apt-get clean

COPY --from=0 /build/hymenaios           /bin/hymenaios
COPY --from=0 /build/config.yml.example  /etc/hymenaios/config.yml
COPY --from=0 /build/LICENSE             /LICENSE

WORKDIR /hymenaios
RUN chown -R nobody:nogroup /etc/hymenaios /hymenaios

USER       nobody
EXPOSE     8080
VOLUME     [ "/hymenaios" ]
ENTRYPOINT [ "/bin/hymenaios" ]

CMD [ "-config.file=/etc/hymenaios/config.yml" ]
