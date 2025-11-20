#!/bin/sh
set -e

# Run database migrations
node_modules/.bin/prisma migrate deploy

# Start the application with exec to properly handle signals
exec node index.js
