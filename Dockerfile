FROM alpine:latest AS webp-builder
WORKDIR /build

RUN apk add --no-cache --update libpng-dev libjpeg-turbo-dev giflib-dev tiff-dev autoconf automake make gcc g++ wget pkgconfig

RUN wget https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.3.2.tar.gz
RUN tar -xvzf libwebp-1.3.2.tar.gz
RUN cd libwebp-1.3.2 && ./configure && make && make install

FROM golang:alpine AS builder

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates
RUN apk add --no-cache --update libpng-dev libjpeg-turbo-dev giflib-dev tiff-dev autoconf automake make gcc g++ wget pkgconfig

ARG BUILD_VERSION

WORKDIR $GOPATH/src/app
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor vendor
COPY config config
COPY app app
COPY stickerUpdater stickerUpdater
COPY logger logger
COPY sender sender
COPY main.go main.go
RUN CGO_ENABLED=0 go build -mod vendor -ldflags="-w -s -X main.version=${BUILD_VERSION}" -trimpath -o /dist/app

FROM scratch
WORKDIR /webp
COPY --from=webp-builder /lib/ld-musl-aarch64.so.1 /lib/ld-musl-aarch64.so.1
COPY --from=webp-builder /usr/local/lib/libwebpdemux.so.2 /usr/local/lib/libwebpdemux.so.2
COPY --from=webp-builder /usr/local/lib/libwebp.so.7 /usr/local/lib/libwebp.so.7
COPY --from=webp-builder /usr/local/lib/libsharpyuv.so.0 /usr/local/lib/libsharpyuv.so.0
COPY --from=webp-builder /usr/lib/libjpeg.so.8 /usr/lib/libjpeg.so.8
COPY --from=webp-builder /usr/lib/libpng16.so.16 /usr/lib/libpng16.so.16
COPY --from=webp-builder /usr/lib/libtiff.so.6 /usr/lib/libtiff.so.6
COPY --from=webp-builder /lib/libz.so.1 /lib/libz.so.1
COPY --from=webp-builder /usr/lib/libzstd.so.1 /usr/lib/libzstd.so.1
COPY --from=webp-builder /usr/local/bin/cwebp /usr/local/bin/cwebp
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /dist /
COPY config.json /config.json
COPY stickerAnon.webp /stickerAnon.webp
ENV SKIP_DOWNLOAD true
ENV VENDOR_PATH /usr/local/bin
ENTRYPOINT ["/app"]

# Build arguments
ARG BUILD_ARCH
ARG BUILD_DATE
ARG BUILD_REF
ARG BUILD_VERSION

# Labels
LABEL \
    io.hass.name="anonstickerbot" \
    io.hass.description="anonstickerbot" \
    io.hass.arch="${BUILD_ARCH}" \
    io.hass.version="${BUILD_VERSION}" \
    io.hass.type="addon" \
    maintainer="ad <github@apatin.ru>" \
    org.label-schema.description="anonstickerbot" \
    org.label-schema.build-date=${BUILD_DATE} \
    org.label-schema.name="anonstickerbot" \
    org.label-schema.schema-version="1.0" \
    org.label-schema.usage="https://gitlab.com/ad/anonstickerbot/-/blob/master/README.md" \
    org.label-schema.vcs-ref=${BUILD_REF} \
    org.label-schema.vcs-url="https://github.com/ad/anonstickerbot/" \
    org.label-schema.vendor="HomeAssistant add-ons by ad"
