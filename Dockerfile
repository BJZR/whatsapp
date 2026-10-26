# Usar la imagen oficial de Go para compilar
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Instalar dependencias del sistema necesarias para CGO y SQLite si fuera necesario
RUN apk add --no-cache git gcc musl-dev

# Copiar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar el binario
RUN CGO_ENABLED=1 GOOS=linux go build -o whatsapp-service .

# Imagen final ligera para ejecución
FROM alpine:latest

WORKDIR /app

# Instalar certificados CA por si acaso
RUN apk add --no-cache ca-certificates tzdata

# Copiar el binario desde el builder
COPY --from=builder /app/whatsapp-service .

# Render asigna dinámicamente un puerto mediante la variable PORT
ENV PORT=8080
EXPOSE 8080

CMD ["./whatsapp-service"]
