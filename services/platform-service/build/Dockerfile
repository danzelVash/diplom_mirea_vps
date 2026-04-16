FROM golang:1.24.4-bookworm AS builder

WORKDIR /src

COPY services/api-gateway /src/services/api-gateway
COPY services/device-service /src/services/device-service
COPY services/edge-bridge-service /src/services/edge-bridge-service
COPY services/scenario-service /src/services/scenario-service
COPY services/voice-service /src/services/voice-service
COPY services/platform-service /src/services/platform-service

WORKDIR /src/services/platform-service
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/platform-service ./cmd

FROM debian:bookworm-slim

RUN useradd --system --create-home --uid 10001 app

COPY --from=builder /out/platform-service /usr/local/bin/platform-service

USER app
WORKDIR /home/app

EXPOSE 8080 9000

ENTRYPOINT ["/usr/local/bin/platform-service"]
