#!/bin/bash
# shellcheck disable=SC2059

set -e

REMOTE_HOST=""
PROJECT_PATH=""

print_usage() {
  echo "Go Game Deploy Script"
  echo "This is a script that I use to deploy my multiplayer web-based game to the server where it is hosted."
  echo "I have included it in my game repository as an example to anyone who may want to use it or gather ideas from it."
  echo "It is not intended to be a general purpose tool, but it may be useful to you."
  echo ""
  echo "Prerequisites:"
  echo "1. You must have SSH access via public key to the remote host."
  echo "2. Your remote host must have unison installed (sudo apt install unison)"
  echo ""
  echo "Usage:"
  echo "You MUST provide the following arguments on the command line:"
  echo "The folder where your Go code project is at:"
  echo "--project-path /mnt/c/Dev/game"
  echo ""
  echo "The game name, which must match your folders."
  echo "--game-name game"
  echo ""
  echo "You MAY also include the following options:"
  echo "The remote host to deploy your code to:"
  echo "--remote-host server.example.com"
  echo "An IP address also works. It is not required if you do not want to deploy to a remote host."
  echo ""
  echo "Example Usage:"
  echo "deployGame.sh --project-path /mnt/c/Dev/game --remote-host server.example.com --game-name game"
}

if [[ $# -eq 0 ]];then
  print_usage
  exit
fi

while test $# -gt 0
do
        case "$1" in
          --remote-host)
            shift
            REMOTE_HOST="$1"
            ;;
          --game-name)
            shift
            GAME_NAME="$1"
            ;;
          --project-path)
            shift
            PROJECT_PATH="$1"
            ;;
          *)
            echo "Invalid argument"
            print_usage
            exit
            ;;
        esac
        shift
done

if [[ ${GAME_NAME} == "" ]] || [[ ${PROJECT_PATH} == "" ]];then
  print_usage
  exit
fi
if ! (command -v go >/dev/null); then
    if [[ -d "/usr/local/go/bin" ]]; then
      export PATH=$PATH:/usr/local/go/bin
    fi
fi

# If Go is not installed download and install the latest version
if ! (command -v go >/dev/null); then
  printf "\n${YELLOW}Looking for Go Language${NC}\n"
  cd /tmp || exit
  # Determine latest version dynamically
  LATEST_VERSION=$(curl -s https://go.dev/dl/ | grep -oP 'go[0-9.]+\.linux-amd64\.tar\.gz' | head -1)
  if [[ "${LATEST_VERSION}" == "" ]]; then
    echo "Could not determine latest Go version"
    exit 1
  fi
  echo "Installing Go version: ${LATEST_VERSION}"
  wget "https://golang.org/dl/${LATEST_VERSION}"
  sudo tar -C /usr/local -xzf "${LATEST_VERSION}"
  rm "${LATEST_VERSION}"
  export PATH=$PATH:/usr/local/go/bin
fi

YELLOW='\033[1;33m'
NC='\033[0m' # NoColor

if ! (command -v zip >/dev/null) || ! (command -v unison >/dev/null); then
  printf "\n${YELLOW}Installing Required Dependencies${NC}\n"
  type -p zip >/dev/null || (sudo apt update && sudo apt install zip -y)
  type -p unison >/dev/null || (sudo apt update && sudo apt install unison -y)
fi

printf "\n${YELLOW}Building Go binaries${NC}"
printf "\n\t${YELLOW}Server${NC}\n"
cd "${PROJECT_PATH}/src" || exit
if [[ -e "${PROJECT_PATH}/src/${GAME_NAME}-server" ]]; then
  rm "${PROJECT_PATH}/src/${GAME_NAME}-server"
fi
go build -o "${PROJECT_PATH}/src/${GAME_NAME}-server" ./server/
printf "\n\t${YELLOW}WASM Client${NC}\n"
GOOS=js GOARCH=wasm go build -o "${PROJECT_PATH}/web/geomyidae.wasm" ./client
GO_ROOT=$(go env GOROOT)
if [[ -e "${PROJECT_PATH}/web/wasm_exec.js" ]]; then
  rm "${PROJECT_PATH}/web/wasm_exec.js"
fi
if [[ -e "$GO_ROOT/misc/wasm/wasm_exec.js" ]]; then
    cp "$GO_ROOT/misc/wasm/wasm_exec.js" "${PROJECT_PATH}/web/"
elif [[ -e "$GO_ROOT/lib/wasm/wasm_exec.js" ]]; then
  cp "$GO_ROOT/lib/wasm/wasm_exec.js" "${PROJECT_PATH}/web/"
else
  echo "Could not find wasm_exec.js in $GO_ROOT/misc/wasm/ or $GO_ROOT/lib/wasm/"
  exit 1
fi

if [[ "${REMOTE_HOST}" != "" ]]; then
  printf "\n${YELLOW}Syncing Builds to Server${NC}"
  printf "\n\t${YELLOW}Syncing Web Content${NC}\n"

  UNISON_ARGUMENTS=()
  UNISON_ARGUMENTS+=("${PROJECT_PATH}/web/")
  UNISON_ARGUMENTS+=("ssh://${USER}@${REMOTE_HOST}//mnt/2000/container-mounts/caddy/site/${GAME_NAME}")
  UNISON_ARGUMENTS+=(-force "${PROJECT_PATH}")
  UNISON_ARGUMENTS+=(-perms)
  UNISON_ARGUMENTS+=(0)
  UNISON_ARGUMENTS+=(-dontchmod)
  UNISON_ARGUMENTS+=(-auto)
  UNISON_ARGUMENTS+=(-batch)
  UNISON_ARGUMENTS+=(-sshcmd "ssh.exe")
  unison "${UNISON_ARGUMENTS[@]}"

  cd "${PROJECT_PATH}/src" || exit
  scp.exe "${GAME_NAME}-server" "${USER}@${REMOTE_HOST}:/mnt/2000/container-mounts/caddy/${GAME_NAME}/"

  printf "\n${YELLOW}Restarting Server${NC}\n"
  # shellcheck disable=SC2029
  ssh.exe "${USER}@${REMOTE_HOST}" "chmod +x /mnt/2000/container-mounts/caddy/${GAME_NAME}/${GAME_NAME}-server;cd /home/${USER}/containers/caddy;docker compose up --detach --build ${GAME_NAME}"
fi
