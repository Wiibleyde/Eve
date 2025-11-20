FROM oven/bun:debian AS builder

WORKDIR /app

RUN apt-get update -y && apt-get install -y openssl

COPY package.json ./
COPY bun.lock ./

RUN bun install --frozen-lockfile

COPY . .

# Ensure Prisma generates binaries for both the builder and the target environment
# Use a dummy DATABASE_URL for build time (prisma.config.ts needs it)
RUN DATABASE_URL="mysql://dummy:dummy@localhost:3306/dummy" bun prisma generate
RUN DATABASE_URL="mysql://dummy:dummy@localhost:3306/dummy" bun run build

FROM node:24.1.0-slim

WORKDIR /app

# Install OpenSSL for Prisma
RUN apt-get update -y && apt-get install -y openssl

# Install Prisma CLI and dependencies needed for prisma.config.ts
RUN npm install -g prisma@7.0.0 && npm install dotenv@17.2.3

# Copy Prisma files for migrations
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/prisma.config.ts ./prisma.config.ts
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