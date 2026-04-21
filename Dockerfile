FROM golang:1.26.1-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS builder
WORKDIR /src
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOFLAGS=-mod=vendor
RUN go build -ldflags="-s -w" -o /out/service-wait ./cmd/

FROM alpine:3.23@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659
RUN adduser -D -g '' appuser
USER appuser
WORKDIR /app

COPY --from=builder /out/ /app/
ENTRYPOINT ["/app/"]