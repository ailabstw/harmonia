#!/bin/sh

DIR=$(basename $PWD)

docker-compose down --volumes --rmi local

docker rm -f test_gitea
docker network rm $DIR

[ -f logs.txt ] && rm logs.txt
[ -d registry_data ] && rm -rf registry_data

rm -rf service_pb2_grpc.py service_pb2.py \
rm -rf edge/service_pb2_grpc.py edge/service_pb2.py
rm -rf aggregator/service_pb2_grpc.py aggregator/service_pb2.py