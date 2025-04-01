# Stage 1: Build
FROM node:22 AS builder
WORKDIR /app

COPY package.json yarn.lock ./

RUN yarn install --production

COPY . .

RUN yarn build

FROM node:22-slim
WORKDIR /app

COPY package.json yarn.lock ./

RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential python3 libopus-dev ffmpeg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN yarn install

COPY . .

RUN yarn build

ENV TZ=Europe/Paris
ENV PATH="/usr/bin:$PATH"

STOPSIGNAL SIGTERM

CMD ["sh", "-c", "yarn prisma generate && yarn prisma migrate deploy && node dist/index.js"]