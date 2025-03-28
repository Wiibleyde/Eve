# Discord bot Dockerfile
FROM node:22
WORKDIR /app

COPY . .

RUN apt-get update && apt-get install -y ffmpeg
RUN yarn install
COPY . .

RUN yarn build

CMD ["sh", "-c", "yarn prisma migrate deploy && yarn start"]