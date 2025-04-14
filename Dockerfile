# Stage 1: Build
FROM node:23-alpine AS builder
WORKDIR /app

# Install build dependencies including Python
RUN apk add --no-cache python3 make g++ tzdata opus opus-dev ffmpeg wget

# Install pnpm and make it available in PATH
RUN wget -qO- https://get.pnpm.io/install.sh | ENV="$HOME/.shrc" SHELL="$(which sh)" sh - && \
    export PATH="/root/.local/share/pnpm:$PATH" && \
    echo 'export PATH="/root/.local/share/pnpm:$PATH"' >> /root/.profile

# Set PATH to include pnpm
ENV PATH="/root/.local/share/pnpm:$PATH"

# Copier d'abord les fichiers de dépendances
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# Make sure Python is properly linked
RUN ln -sf /usr/bin/python3 /usr/bin/python && \
    # Installer les dépendances de développement pour le build
    pnpm install

# Copier le reste du code source
COPY . .

# Générer les fichiers Prisma et construire l'application
RUN pnpm prisma generate && \
    pnpm build

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
    ln -sf /usr/bin/python3 /usr/bin/python

# Install pnpm
RUN wget -qO- https://get.pnpm.io/install.sh | ENV="$HOME/.shrc" SHELL="$(which sh)" sh - && \
    export PATH="/root/.local/share/pnpm:$PATH" && \
    echo 'export PATH="/root/.local/share/pnpm:$PATH"' >> /root/.profile

# Set PATH to include pnpm
ENV PATH="/root/.local/share/pnpm:$PATH"

# Copier les fichiers de dépendances
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# Installer uniquement les dépendances de production avec cache cleanup
# RUN yarn install --production --frozen-lockfile && \
#     yarn cache clean
RUN pnpm install --prod && \
    # Supprimer les fichiers de cache pour réduire la taille de l'image
    rm -rf /root/.cache

# Copier les fichiers générés depuis l'étape de build
COPY --from=builder /app/dist ./dist
# COPY --from=builder /app/node_modules/.prisma ./node_modules/.prisma
COPY --from=builder /app/prisma ./prisma
COPY --from=builder /app/assets ./assets

# Copy node_modules instead of just .prisma directory
COPY --from=builder /app/node_modules ./node_modules

STOPSIGNAL SIGTERM

CMD ["sh", "-c", "pnpm prisma migrate deploy && node dist/index.js"]