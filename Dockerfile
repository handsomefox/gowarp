FROM golang:1.18-alpine AS builder

RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

ENV USER=appuser
ENV UID=1000
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR $GOPATH/src/mypackage/myapp/
COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    GOMODCACHE=/go/pkg/mod go mod download all
RUN go mod verify

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build --mount=type=cache,target=/go/pkg/mod \
    GOCACHE=/root/.cache/go-build GOMODCACHE=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' -a \
    -o /go/bin/app ./cmd/gowarp/.

FROM scratch

ENV DOCKERIZED=true

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/app /go/bin/app

USER appuser:appuser

ENTRYPOINT ["/go/bin/app"]
