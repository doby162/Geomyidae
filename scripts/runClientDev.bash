#!/usr/bin/env bash

set -e

echo "This will run the client game in a loop so that it restarts if it is closed or crashes or if the server restarts."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../src" || exit 1
while true; do
    go run ./client
done
