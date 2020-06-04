#!/bin/sh

DIR=$(basename $PWD)
PROJ_ROOT=$(git rev-parse --show-toplevel)

spin () {
    i=1

    while [ $i -le $(( $1 * 2 )) ]
    do
        printf "."
        sleep .5s
        i=$((i+1))
    done
    printf '\n'
}

echo "Cleaning previous outputs..."
./test_cleanup.sh

echo "Creating network"
docker network create ${DIR}

echo "Creating gitea"
docker run -d \
    --volume ${PWD}/registry_data:/data \
    --network ${DIR} \
    --env INSTALL_LOCK=true \
    --env LFS_START_SERVER=true \
    --env SSH_DOMAIN=test_gitea \
    --env ROOT_URL=http://test_gitea:3000/ \
    -p 3000:3000 \
    --name test_gitea \
    gitea/gitea

echo "Warm up gitea..."
spin 5

echo "Creating gitea admin"
docker exec test_gitea gitea admin create-user --admin \
    --username harmonia_admin \
    --password password \
    --email admin@admin.com

echo "Creating repositories and webhooks"
docker run --rm \
    --volume ${PWD}:/setup \
    --network ${DIR} \
    python bash -c "pip3 install requests==2.21.* && python3 /setup/setup.py init"

echo "Building Python Protocol"
docker run --rm \
    -v $PROJ_ROOT:/proj_root \
    -w /proj_root \
    python /proj_root/test/integration-test/protos_build.sh
cp service_pb2_grpc.py service_pb2.py edge/
cp service_pb2_grpc.py service_pb2.py aggregator/

echo "Start Integration Test"
docker-compose up -d --build

until [ "`docker inspect -f {{.State.Running}} ${DIR}_aggregator-app_1`"=="true" ]; do
    sleep 1;
done;
until [ "`docker inspect -f {{.State.Running}} ${DIR}_edge-app_1`"=="true" ]; do
    sleep 1;
done;
sleep 1;

docker run --rm \
    --volume ${PWD}:/setup \
    --network ${DIR} \
    python bash -c "pip3 install requests==2.21.* && python3 /setup/setup.py set-plan '{\"roundCount\": 2, \"edgeCount\": 1, \"epochCount\": 1}'"

SEC=20
[ ! -z "$GITLAB_CI" ] && SEC=240
echo "sleeping ${SEC} seconds to wait the harmonia system finished one round"
spin $SEC
echo "stopping docker-compose"
docker-compose stop
docker-compose logs --no-color > logs.txt
python3 parsing_log.py logs.txt
