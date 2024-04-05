FROM golang:alpine AS builder

RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates
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
COPY config.json /config.json
COPY stickerAnon.webp /stickerAnon.webp
RUN go version
RUN CGO_ENABLED=1 go build -mod vendor -ldflags="-w -s -X main.version=${BUILD_VERSION}" -o /go/bin/app .

FROM scratch
WORKDIR /app/
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY config.json /config.json
COPY --from=builder /go/bin/app /go/bin/app
ENTRYPOINT ["/go/bin/app"]

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
