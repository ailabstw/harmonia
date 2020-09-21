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

echo "Start Integration Test"
docker-compose up -d --build

until [ "`docker inspect -f {{.State.Running}} ${DIR}_aggregator-app_1`"=="true" ]; do
    sleep 1;
done;
until [ "`docker inspect -f {{.State.Running}} ${DIR}_push-edge-app_1`"=="true" ]; do
    sleep 1;
done;
until [ "`docker inspect -f {{.State.Running}} ${DIR}_pull-edge-app_1`"=="true" ]; do
    sleep 1;
done;
sleep 1;

docker run --rm \
    --volume ${PWD}:/setup \
    --network ${DIR} \
    python bash -c "pip3 install requests==2.21.* && python3 /setup/setup.py set-plan '{\"round\": 2, \"edge\": 3, \"EpR\": 3, \"timeout\": 5}'"

time docker wait \
    ${DIR}_aggregator-app_1 ${DIR}_push-edge-app_1 ${DIR}_pull-edge-app_1 \
    ${DIR}_aggregator-operator_1 ${DIR}_push-edge-operator_1 ${DIR}_pull-edge-operator_1    

docker-compose logs --no-color > logs.txt
python3 parsing_log.py logs.txt
