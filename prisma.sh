#!/bin/bash
# This script is used to generate the Prisma client for Go and run the Prisma CLI commands.

if [ $# -eq 0 ]; then
  echo "Usage: $0 <command>"
  echo "Example: $0 generate"
  echo "Available commands are the same as prisma-client-go commands."
  exit 1
fi

# Pass all arguments to prisma-client-go
go run github.com/steebchen/prisma-client-go "$@"