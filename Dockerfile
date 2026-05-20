FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o notes-api ./cmd/api

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/notes-api .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./notes-api"]