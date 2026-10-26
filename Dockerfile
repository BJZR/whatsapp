FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev
COPY go.mod ./
RUN go mod download || true
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux go build -o whatsapp-service .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /app/whatsapp-service .
ENV PORT=8080
EXPOSE 8080
CMD ["./whatsapp-service"]
