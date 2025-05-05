FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run github.com/steebchen/prisma-client-go prefetch
RUN go run github.com/steebchen/prisma-client-go generate

RUN CGO_ENABLED=0 GOOS=linux go build -o app .

# Image finale : Node + Prisma CLI + app Go + certificats
FROM node:20-slim

WORKDIR /app

# 🛠️ Ajout de OpenSSL + certificats CA
RUN apt-get update && apt-get install -y \
    openssl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Prisma CLI (pour migrate deploy)
COPY package.json ./
RUN npm install --omit=dev

# Prisma schema
COPY prisma ./prisma

# App Go + assets
COPY --from=builder /app/app .
COPY --from=builder /app/assets ./assets

# Script d'entrée
COPY docker-entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["./entrypoint.sh"]