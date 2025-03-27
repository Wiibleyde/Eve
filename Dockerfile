# Discord bot Dockerfile
FROM node:22
WORKDIR /app

COPY . .

RUN yarn install
COPY . .

RUN yarn build

CMD ["sh", "-c", "yarn prisma migrate deploy && yarn start"]