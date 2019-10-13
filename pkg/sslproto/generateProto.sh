#!/bin/sh
go get -u github.com/golang/protobuf/protoc-gen-go

packageName=${PWD##*/}
protoc --go_out=import_path="${packageName}:." ./*.proto
