#!/bin/bash
DIR=$(dirname $0)

python3 -m pip install --user grpcio-tools
python3 -m grpc_tools.protoc \
        -I src/protos \
        --python_out=./test/integration-test \
        --grpc_python_out=./test/integration-test \
        src/protos/service.proto