# Stage 1: Build
FROM debian:bullseye AS builder
WORKDIR /app

# Installer Node.js et Yarn
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g yarn \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY package.json yarn.lock ./

RUN yarn install --production

COPY . .

RUN yarn build

# Stage 2: Runtime
FROM node:22
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    libopus-dev \
    python3 \
    curl \
    && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/assets ./assets

ENV TZ=France/Paris

STOPSIGNAL SIGTERM

CMD ["sh", "-c", "yarn prisma generate && yarn prisma migrate deploy && node dist/index.js"]