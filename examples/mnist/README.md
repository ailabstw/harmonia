# MNIST

An example of Harmonia with 2-edge MNIST

## Preliminary
1. Build `operator`
    `$ cd ../../ & make all `
2. Build python grpc packages in `../../src/protos`
    ```
    docker run --rm \
        -v $(PWD):/protos \
        -w /protos \
        python bash -c "python3 -m pip install --user grpcio-tools && python3 -m grpc_tools.protoc -I . --python_out=. --grpc_python_out=. service.proto"
    ```
And move `service_pb2.py`, `service_pb2_grpc.py` into `aggregator` and `edge` to build aplication images.
3. Start a Gitea service
    `$ kubectl apply -f registry.yml`

## Steps

A step by step series of examples explain how to get a development/production env running

1. Build `application` images
    `$ cd aggregator; docker build . --tag <registry>/mnist-aggregator`
    `$ cd edge; docker build . --tag <registry>/mnist-edge`

2. Push images
    `$ docker tag operator <registry>/operator && docker push <registry>/operator`
    `$ docker push <registry>/mnist-aggregator`
    `$ docker push <registry>/mnist-edge`

3. Apply configs
    `$ kubectl apply -f configs.yml`

4. Setup accounts, repositories, and webhooks by Gitea UI reference to `configs.yml`
    `global-model`
    `local-model1`
    `local-model2`
    `train-plan`

5. Apply mnist deployment
    `$ kubectl apply -f mnist-deployment.yml`

6. Trigger training by updating train plan `plan.json` and pushing to the `train-plan` repository
```json
{
    "roundCount": 10,
    "edgeCount": 2
}
```
