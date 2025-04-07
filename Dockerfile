# Stage 1: Build
FROM node:22 AS builder
WORKDIR /app

# Copier d'abord les fichiers de dépendances
COPY package.json yarn.lock ./

# Installer les dépendances de développement pour le build
RUN yarn install

# Copier le reste du code source
COPY . .

# Générer les fichiers Prisma et construire l'application
RUN yarn prisma generate && yarn build

# Stage 2: Production
FROM node:22-slim
WORKDIR /app

# Définir le fuseau horaire
ENV TZ=Europe/Paris
ENV PATH="/usr/bin:$PATH"

# Installer les dépendances systèmes nécessaires
RUN apt-get update && \
    apt-get install -y --no-install-recommends build-essential python3 libopus-dev ffmpeg && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

# Copier les fichiers de dépendances
COPY package.json yarn.lock ./

# Installer uniquement les dépendances de production
RUN yarn install --production

# Copier les fichiers générés depuis l'étape de build
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules/.prisma ./node_modules/.prisma
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/assets ./assets

STOPSIGNAL SIGTERM

CMD ["sh", "-c", "yarn prisma migrate deploy && node dist/index.js"]