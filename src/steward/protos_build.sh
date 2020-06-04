#!/bin/bash
DIR=$(dirname $0)

apt-get update && apt-get install -y protobuf-compiler

go get github.com/golang/protobuf/protoc-gen-go

mkdir -p ${DIR}/protos
protoc -I ${DIR}/../protos/ --go_out=${DIR}/protos ${DIR}/../protos/service.proto
