#!/bin/sh
set -e

echo "🔄 Déploiement des migrations Prisma..."
npx prisma migrate deploy

echo "🚀 Lancement de l'application Go..."
exec ./app
