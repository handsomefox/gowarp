FROM golang:1.18-alpine AS builder

RUN apk update
RUN apk add --no-cache git
RUN apk add -U --no-cache ca-certificates && update-ca-certificates

COPY . "/build/"
WORKDIR "/build"
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"

FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/gowarp /app/

WORKDIR /app
EXPOSE $PORT
ENTRYPOINT ["./gowarp"]
