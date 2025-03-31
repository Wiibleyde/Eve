# Stage 1: Build
FROM node:22 AS builder
WORKDIR /app

COPY package.json yarn.lock ./

RUN yarn install --production

COPY . .

RUN yarn build

FROM node:22-slim
WORKDIR /app

RUN apt-get update && apt-get install -y ffmpeg && apt-get clean && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/assets ./assets

ENV TZ=France/Paris

STOPSIGNAL SIGTERM

CMD ["sh", "-c", "yarn prisma generate && yarn prisma migrate deploy && node dist/index.js"]