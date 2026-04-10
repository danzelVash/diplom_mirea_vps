FROM golang:1.24-bookworm AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /core-backend ./cmd/core

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /core-backend /usr/local/bin/core-backend
EXPOSE 8080
ENTRYPOINT ["core-backend"]
CMD ["--listen", ":8080", "--data-dir", "/app/data"]
