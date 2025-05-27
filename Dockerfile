FROM oven/bun:latest AS builder

WORKDIR /app

COPY package.json ./
COPY bun.lock ./

RUN bun install --frozen-lockfile

COPY . .

# Ensure Prisma generates binaries for both the builder and the target environment
RUN bun prisma generate
RUN bun run build

FROM node:24.1.0-slim

WORKDIR /app

# Install OpenSSL for Prisma
RUN apt-get update -y && apt-get install -y openssl

COPY --from=builder /app/dist ./

# For dev
# COPY --from=builder /app/.env ./

CMD [ "node", "index.js" ]