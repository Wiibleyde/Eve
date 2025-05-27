FROM oven/bun:debian AS builder

WORKDIR /app

RUN apt-get update -y && apt-get install -y openssl

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

# Copy Prisma files for migrations
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/node_modules/.prisma ./node_modules/.prisma
COPY --from=builder /app/node_modules/@prisma ./node_modules/@prisma

COPY --from=builder /app/dist ./

# For dev
# COPY --from=builder /app/.env ./

# Copy the entrypoint script
COPY --from=builder /app/docker-entrypoint.sh ./
RUN chmod +x /app/docker-entrypoint.sh

# Use ENTRYPOINT instead of CMD for proper signal handling
ENTRYPOINT ["/app/docker-entrypoint.sh"]