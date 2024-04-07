FROM golang:alpine AS builder

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates

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
