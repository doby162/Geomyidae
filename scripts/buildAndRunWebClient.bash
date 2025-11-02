#!/usr/bin/env bash
# *** NOTICE: THIS WAS AI GENERATED! ***
# buildAndRunWebClient.bash
# Minimal portable bash conversion template for a PowerShell buildAndRunWebClient.ps1
# Edit the commands in build_project() and run_project() to match your original PS1 logic.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Default: assume web client is in a sibling "web-client" directory. Change as needed.
WEB_CLIENT_DIR="$SCRIPT_DIR/../web-client"
# PID of background run (if any)
RUN_PID=""

usage() {
    cat <<EOF
Usage: $(basename "$0") [--build] [--run] [--dir PATH] [--help]

Options:
    --build       Only build the web client
    --run         Only run the web client
    --dir PATH    Path to web client project (default: $WEB_CLIENT_DIR)
    --help        Show this help
If neither --build nor --run is provided, the script will build then run.
EOF
}

check_command() {
    command -v "$1" >/dev/null 2>&1 || {
        echo "Required command '$1' not found in PATH. Aborting." >&2
        exit 2
    }
}

build_project() {
    echo "Building web client in: $WEB_CLIENT_DIR"
    if [ -f "$WEB_CLIENT_DIR/package.json" ]; then
        check_command npm
        (cd "$WEB_CLIENT_DIR" && npm install)
        # replace with your build command (npm run build, yarn build, dotnet build, etc.)
        (cd "$WEB_CLIENT_DIR" && npm run build)
    elif [ -f "$WEB_CLIENT_DIR/Project.csproj" ] || ls "$WEB_CLIENT_DIR"/*.csproj >/dev/null 2>&1; then
        check_command dotnet
        (cd "$WEB_CLIENT_DIR" && dotnet restore && dotnet build --configuration Release)
    else
        echo "No recognized project file (package.json or .csproj) found in $WEB_CLIENT_DIR" >&2
        exit 3
    fi
    echo "Build complete."
}

run_project() {
    echo "Running web client from: $WEB_CLIENT_DIR"
    if [ -f "$WEB_CLIENT_DIR/package.json" ]; then
        check_command npm
        # Start in foreground; if you want background, append & and capture PID
        (cd "$WEB_CLIENT_DIR" && npm start) &
        RUN_PID=$!
        echo "Web client started (PID=$RUN_PID)."
        # wait for process to exit so the script doesn't exit immediately
        wait "$RUN_PID"
    elif [ -f "$WEB_CLIENT_DIR/Project.csproj" ] || ls "$WEB_CLIENT_DIR"/*.csproj >/dev/null 2>&1; then
        check_command dotnet
        (cd "$WEB_CLIENT_DIR" && dotnet run --project .) &
        RUN_PID=$!
        echo "Web client started (PID=$RUN_PID)."
        wait "$RUN_PID"
    else
        echo "No recognized project file (package.json or .csproj) found in $WEB_CLIENT_DIR" >&2
        exit 3
    fi
}

trap 'on_exit' EXIT
on_exit() {
    if [ -n "${RUN_PID:-}" ] && ps -p "$RUN_PID" >/dev/null 2>&1; then
        echo "Stopping background process $RUN_PID"
        kill "$RUN_PID" || true
    fi
}

# Parse args
DO_BUILD=false
DO_RUN=false

while [ $# -gt 0 ]; do
    case "$1" in
        --build) DO_BUILD=true; shift ;;
        --run) DO_RUN=true; shift ;;
        --dir) WEB_CLIENT_DIR="$2"; shift 2 ;;
        --help|-h) usage; exit 0 ;;
        *) echo "Unknown arg: $1" >&2; usage; exit 1 ;;
    esac
done

# Default behavior: build then run
if ! $DO_BUILD && ! $DO_RUN; then
    DO_BUILD=true
    DO_RUN=true
fi

# Ensure path exists
if [ ! -d "$WEB_CLIENT_DIR" ]; then
    echo "Web client directory does not exist: $WEB_CLIENT_DIR" >&2
    exit 4
fi

if $DO_BUILD; then
    build_project
fi

if $DO_RUN; then
    run_project
fi