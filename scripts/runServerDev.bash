#!/usr/bin/env bash

set -e

echo "This will run the server and restart it if any code is changed."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../src" || exit 1
air
