ARG VERSION=0.2.8

FROM alpine:3 as downloader

ARG VERSION
ARG TARGETPLATFORM

WORKDIR /tmp

# install cosign
COPY --from=gcr.io/projectsigstore/cosign:v1.6.0@sha256:319eab5b1563b8e1c02301ea4e7156556c9170edaeb7ce7791b780e4ee00e0bc /ko-app/cosign /usr/local/bin/cosign

RUN \
  case ${TARGETPLATFORM} in \
    "linux/amd64") DOWNLOAD_ARCH="linux-amd64"  ;; \
    "linux/arm64") DOWNLOAD_ARCH="linux-arm64"  ;; \
  esac && \
  apk add --no-cache curl upx && \
  curl -sLO https://github.com/elastic/harp/releases/download/v${VERSION}/harp-${DOWNLOAD_ARCH}.tar.gz && \
  curl -sLO https://github.com/elastic/harp/releases/download/v${VERSION}/harp-${DOWNLOAD_ARCH}.tar.gz.sig && \
  curl -sLO https://raw.githubusercontent.com/elastic/harp/v${VERSION}/build/artifact/cosign.pub && \
  cosign verify-blob --key /tmp/cosign.pub --signature harp-${DOWNLOAD_ARCH}.tar.gz.sig harp-${DOWNLOAD_ARCH}.tar.gz && \
  tar -vxf harp-${DOWNLOAD_ARCH}.tar.gz && \
  mv /tmp/harp-${DOWNLOAD_ARCH} /tmp/harp && \
  upx -9 /tmp/harp && \
  chmod +x /tmp/harp

FROM alpine:3

ARG VERSION

RUN apk update --no-cache && \
    apk add --no-cache ca-certificates && \
    rm -rf /var/cache/apk/*

COPY --from=downloader /tmp/harp /usr/bin/harp

ENTRYPOINT [ "/usr/bin/harp" ]
