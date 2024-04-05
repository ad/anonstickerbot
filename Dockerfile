FROM golang:alpine AS builder

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates
RUN apk add build-base
RUN apk add gcc g++ libwebp-dev

ARG BUILD_VERSION

WORKDIR $GOPATH/src/app
COPY go.mod go.mod
COPY go.sum go.sum
COPY vendor vendor
COPY config config
COPY logger logger
COPY sender sender
COPY main.go main.go
COPY draw.go draw.go
COPY data.go data.go
RUN CGO_ENABLED=1 go build -mod vendor -ldflags="-w -s -X main.version=${BUILD_VERSION}" -trimpath -o /dist/app
RUN ldd /dist/app | tr -s [:blank:] '\n' | grep ^/ | xargs -I % install -D % /dist/%
RUN ls -la /dist/lib/
RUN ln -s ld-musl-arm64.so.1 /dist/lib/libc.musl-arm64.so.1

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /dist /
COPY config.json /config.json
COPY stickerAnon.webp /stickerAnon.webp
ENTRYPOINT ["/app"]

#
# LABEL target docker image
#

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
