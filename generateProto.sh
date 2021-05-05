#!/bin/bash

# Fail on errors
set -e
# Print commands
set -x

# Update to latest protobuf compiler
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Generate all protobuf code
protoc -I"./proto" -I"$GOPATH/src" --go_out="$GOPATH/src" proto/*.proto
