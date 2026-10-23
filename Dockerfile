FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/shieldgate ./cmd/shieldgate/main.go

FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /app/bin/shieldgate /app/shieldgate
COPY --from=builder /app/web /app/web
EXPOSE 8080
CMD ["/app/shieldgate"]