#!/bin/bash
set -euo pipefail

# determine current script directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
PROJECT_DIR="${SCRIPT_DIR}/.."
cd "${PROJECT_DIR}"

PB_VERSION=3.15.8
PB_GO_VERSION=$(go list -m all | grep google.golang.org/protobuf | awk '{print $2}')

# Create a local bin folder
LOCAL_DIR=".local"
mkdir -p "${LOCAL_DIR}"

# install a specific version of protoc
PB_REL="https://github.com/protocolbuffers/protobuf/releases"
if ! protoc --version | grep "${PB_VERSION}" >/dev/null; then
  if [[ ! -f "${LOCAL_DIR}/bin/protoc" ]]; then
    curl -sLO "$PB_REL/download/v${PB_VERSION}/protoc-${PB_VERSION}-linux-x86_64.zip"
    unzip "protoc-${PB_VERSION}-linux-x86_64.zip" -d "${LOCAL_DIR}"
    rm "protoc-${PB_VERSION}-linux-x86_64.zip"
  fi
  export PATH="${LOCAL_DIR}/bin:$PATH"
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

# Generate all protobuf code
protoc -I"./proto" -I"$GOPATH/src" --go_out="$GOPATH/src" proto/*.proto
