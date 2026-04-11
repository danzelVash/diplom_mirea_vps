FROM golang:1.24.4-alpine

WORKDIR /app
COPY . .

WORKDIR /app/services/edge-bridge-service
CMD ["go", "run", "./cmd"]
