#!/bin/sh
set -e

# Run database migrations
prisma migrate deploy

# Start the application with exec to properly handle signals
exec node index.js
