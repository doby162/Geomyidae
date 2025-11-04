#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../src" || exit 1
GOOS=js GOARCH=wasm go build -o ../web/geomyidae.wasm ./client
cd "$SCRIPT_DIR/../web" || exit 1
GO_ROOT=$(go env GOROOT)
cp "$GO_ROOT/misc/wasm/wasm_exec.js" .
echo "Serving web client at http://localhost:8081"
python3 -m http.server 8081
