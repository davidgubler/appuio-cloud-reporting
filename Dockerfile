FROM docker.io/library/alpine:3.15 as runtime

RUN \
  apk add --update --no-cache \
    bash \
    curl \
    ca-certificates \
    tzdata

ENTRYPOINT ["appuio-cloud-reporting"]
COPY appuio-cloud-reporting /usr/bin/

USER 65536:0