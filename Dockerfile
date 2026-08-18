FROM --platform=$BUILDPLATFORM golang:1.25.5-alpine AS builder

WORKDIR /app

ARG TARGETOS=linux
ARG TARGETARCH

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o tracker-bot ./cmd/tracker-bot
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o migrator ./cmd/migrator

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata && adduser -D -H appuser

COPY --from=builder /app/tracker-bot /app/tracker-bot
COPY --from=builder /app/migrator /app/migrator

COPY migrations /app/migrations

USER appuser

CMD ["/app/tracker-bot"]
