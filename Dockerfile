ARG GO_VERSION="1.18.1"
ARG DEBIAN_VERSION="bullseye"
FROM golang:${GO_VERSION}-${DEBIAN_VERSION}
LABEL maintainer="The Hymenaios Authors <developers@hymenaios.io>"

ARG  ARCH="amd64"
ARG  OS="linux"
COPY config.yml.example              /etc/hymenaios/config.yml
COPY .build/${OS}/${ARCH}/hymenaios  /bin/hymenaios
COPY LICENSE                         /LICENSE

WORKDIR /hymenaios
RUN chown -R nobody:nogroup /etc/hymenaios /hymenaios

USER       nobody
EXPOSE     8080
VOLUME     [ "/hymenaios" ]
ENTRYPOINT [ "/bin/hymenaios" ]

CMD [ "-config.file=/etc/hymenaios/config.yml" ]
