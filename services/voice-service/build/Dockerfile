FROM golang:1.24.4-alpine

WORKDIR /app
COPY . .

WORKDIR /app/services/voice-service
CMD ["go", "run", "./cmd"]
