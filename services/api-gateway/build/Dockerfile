FROM golang:1.24.4-alpine

WORKDIR /app
COPY . .

WORKDIR /app/services/api-gateway
CMD ["go", "run", "./cmd"]
