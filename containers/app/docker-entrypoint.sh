#!/bin/sh

set -e

echo "Running $1..."

if [ "$1" = 'api' ]; then
    exec /app/api
elif [ "$1" = 'worker' ]; then
    exec /app/worker
fi
