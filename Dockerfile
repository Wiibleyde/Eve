# Stage 1: Build
FROM node:23-alpine AS builder
WORKDIR /app

# Copier d'abord les fichiers de dépendances
COPY package.json yarn.lock ./

# Installer les dépendances de développement pour le build
RUN yarn install

# Copier le reste du code source
COPY . .

# Générer les fichiers Prisma et construire l'application
RUN yarn prisma generate && \
    yarn build && \
    yarn cache clean

# Stage 2: Production
FROM node:23-alpine AS production
WORKDIR /app

# Définir le fuseau horaire
ENV TZ=Europe/Paris
ENV NODE_ENV=production

# Installer uniquement les dépendances nécessaires
RUN apk --no-cache add tzdata ffmpeg opus python3 g++ make && \
    cp /usr/share/zoneinfo/Europe/Paris /etc/localtime && \
    echo "Europe/Paris" > /etc/timezone && \
    apk del tzdata

# Copier les fichiers de dépendances
COPY package.json yarn.lock ./

# Installer uniquement les dépendances de production avec cache cleanup
RUN yarn install --production --frozen-lockfile && \
    yarn cache clean

# Copier les fichiers générés depuis l'étape de build
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules/.prisma ./node_modules/.prisma
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/assets ./assets

STOPSIGNAL SIGTERM

CMD ["sh", "-c", "yarn prisma migrate deploy && node dist/index.js"]