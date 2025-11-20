#!/bin/sh
set -e

# Run database migrations using the Prisma CLI directly
node node_modules/prisma/build/index.js migrate deploy

# Start the application with exec to properly handle signals
exec node index.js
