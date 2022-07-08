########
# BASE #
########
ARG GO_VERSION="1.18.3"
ARG DEBIAN_VERSION="bullseye"
FROM golang:${GO_VERSION}-${DEBIAN_VERSION}

COPY . /build/
WORKDIR /build/

ARG BUILD_VERSION="development"
RUN make BUILD_VERSION=${BUILD_VERSION} go-build
RUN chmod 755 /build/argus


#############
# ARGUS #
#############
ARG DEBIAN_VERSION="bullseye"
FROM debian:${DEBIAN_VERSION}-slim
LABEL maintainer="The Argus Authors <developers@release-argus.io>"
RUN \
    apt-get update && \
    apt-get install ca-certificates -y && \
    apt-get clean

COPY --from=0 /build/argus               /bin/argus
COPY --from=0 /build/config.yml.example  /etc/argus/config.yml
COPY --from=0 /build/LICENSE             /LICENSE

WORKDIR /argus
RUN chown -R nobody:nogroup /etc/argus /argus

USER       nobody
EXPOSE     8080
VOLUME     [ "/argus" ]
ENTRYPOINT [ "/bin/argus" ]

CMD [ "-config.file=/etc/argus/config.yml" ]
