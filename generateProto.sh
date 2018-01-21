#!/bin/sh
go get -u github.com/golang/protobuf/protoc-gen-go

cd reader
protoc --go_out=import_path=main:. *.proto