# Docker Deployment

For simplification, we simulate real network by a docker network (`mnist`) , and each `docker-compose` within subfolder (`aggregator`, `edge1`, `edge2` and `logserver`) is seen as a isolated computation entity for proof of concept.

## Steps

1. Create `mnist` docker network
    ```bash
    docker network create mnist
    ```

2. Create a Gitea instance.
    ```bash
    docker run -d \
        --env LFS_START_SERVER=true \
        --env INSTALL_LOCK=true \
        --env ROOT_URL=http://gitea:3000 \
        --publish 3000:3000 \
        --network mnist \
        --name gitea \
        gitea/gitea

    # Notes:
    # `LFS_START_SERVER` enables git lfs in Gitea which is required by Harmonia.
    # `ROOT_URL` shuould be set the same as container name to ensure git operation work.
    ```

3. Setup Gitea for the MNIST example:  
    ```bash
    docker cp ./gitea_setup.sh gitea:/gitea_setup.sh
    docker exec gitea bash /gitea_setup.sh
    ```

    Including creates
    * admin account: `gitea` (password: `password`)
    * user accounts: `aggregator` `edge1` `edge2` `logserver`
    * repositories: `train-plan` `global-model` `local-model1` `local-model2`
    * repository permissions: TODO
    * webhooks:
        * `train-plan` to `http://aggregator:9080` `http://edge1:9080` `http://edge2:9080`
        * `global-model` to `http://edge1:9080` `http://edge2:9080` `http://logserver:9080`
        * `local-model1` to `http://aggregator:9080` `http://logserver:9080`
        * `local-model2` to `http://aggregator:9080` `http://logserver:9080`

4. Push the pretrained model to `global-model`.
    ```bash
    docker network connect bridge gitea

    git clone http://gitea@localhost:3000/gitea/global-model.git
    pushd global-model

    git commit -m "pretrained model" --allow-empty
    git push origin master

    popd
    rm -rf global-model
    ```

5. Deploy every FL participants by `docker-compose up` in each folder
    ```bash
    pushd aggregator
    docker-compose up -d
    popd
    
    pushd edge1
    docker-compose up -d
    popd
    
    pushd edge2
    docker-compose up -d
    popd

    pushd logserver
    docker-compose up -d
    popd
    ```

6. Push a train plan to trigger federated MNIST.
    ```bash
    # bash
    docker network connect bridge gitea

    git clone http://gitea@localhost:3000/gitea/train-plan.git
    pushd train-plan

    cat > plan.json << EOF
    {
        "name": "MNIST",
        "round": 10,
        "edge": 2,
        "EpR": 1,
        "timeout": 86400,
        "pretrainedModel": "master"
    }
    EOF

    git add plan.json
    git commit -m "train plan commit"
    git push origin master

    popd
    rm -rf train-plan
    ```
    You can check FL result within Gitea web UI (`http://localhost:3000`) or Tensorboard UI (`http://localhost:6006`).