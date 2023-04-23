#!/bin/bash
set -euo pipefail

# determine current script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
PROJECT_DIR="${SCRIPT_DIR}/.."
cd "${PROJECT_DIR}"

PB_VERSION=22.3
PB_GO_VERSION=$(go list -m all | grep google.golang.org/protobuf | awk '{print $2}')

# Create a local bin folder
LOCAL_DIR=".local/protoc"
mkdir -p "${LOCAL_DIR}"

# install a specific version of protoc
export PATH="${LOCAL_DIR}/bin:$PATH"
if ! protoc --version | grep "${PB_VERSION}" >/dev/null; then
  if [[ -d "${LOCAL_DIR}" ]]; then
    rm -rf "${LOCAL_DIR}"
  fi
  curl -sLO "https://github.com/protocolbuffers/protobuf/releases/download/v${PB_VERSION}/protoc-${PB_VERSION}-linux-x86_64.zip"
  unzip "protoc-${PB_VERSION}-linux-x86_64.zip" -d "${LOCAL_DIR}"
  rm "protoc-${PB_VERSION}-linux-x86_64.zip"
fi

# install a specific version of protoc-gen-go
go install "google.golang.org/protobuf/...@${PB_GO_VERSION}"

###

if ! protoc --version | grep "${PB_VERSION}"; then
  echo "protoc version is not ${PB_VERSION}"
  exit 1
fi

if ! protoc-gen-go --version | grep "${PB_GO_VERSION}"; then
  echo "protoc-gen-go version is not ${PB_GO_VERSION}"
  exit 1
fi

###

# Print commands
set -x

# Generate Go code
protoc -I"./proto" -I"$GOPATH/src" --go_out=. --go_opt=module=github.com/RoboCup-SSL/ssl-go-tools proto/*.proto

