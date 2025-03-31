# Stage 1: Build
FROM node:22 AS builder
WORKDIR /app

COPY package.json yarn.lock ./

RUN yarn install --production

COPY . .

RUN yarn build

# Stage 2: Runtime
FROM debian:11
WORKDIR /app

# Installer les dépendances nécessaires
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ffmpeg \
    libopus-dev \
    python3 \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

# Copy Node.js, npm, and yarn from the builder stage
COPY --from=builder /usr/local/bin/node /usr/local/bin/node
COPY --from=builder /usr/local/lib/node_modules /usr/local/lib/node_modules
COPY --from=builder /usr/local/bin/npm /usr/local/bin/npm

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/assets ./assets

ENV TZ=France/Paris

STOPSIGNAL SIGTERM

CMD ["sh", "-c", "yarn prisma generate && yarn prisma migrate deploy && node dist/index.js"]