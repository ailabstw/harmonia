# Kubernetes Deployment

## Steps

In a kubernetes cluster, there is an registry for deployments get images. Replace all `<image_registry>` with your cluster setting in following commands and Harmonia configuration.

1. Push Harmonia and MNIST assosicated images before deployment
    ```bash
    docker tag harmonia/operator <image_registry>/harmonia/operator
    docker tag harmonia/fedavg <image_registry>/harmonia/fedavg
    docker tag harmonia/logserver <image_registry>/harmonia/logserver
    docker tag mnist_edge <image_registry>/mnist_edge

    docker push <image_registry>/harmonia/operator
    docker push <image_registry>/harmonia/fedavg
    docker push <image_registry>/harmonia/logserver
    docker push <image_registry>/mnist_edge
    ```

2. Start a Gitea service
    ```bash
    kubectl apply -f registry.yml
    ```
3. Setup accounts, repositories, and webhooks by Gitea UI or following script
    ```bash
    kubectl cp gitea_setup.sh $(kubectl get pods -l app=gitea -o template --template="{{(index .items 0).metadata.name }}"):/
    kubectl exec $(kubectl get pods -l app=gitea -o template --template="{{(index .items 0).metadata.name }}") bash /gitea_setup.sh
    ```
4. Push the pretrained model
    ```bash
    # Set port-forward to access gitea by `localhost`
    kubectl port-forward service/harmonia-gitea 3000
    ```
    ```bash
    git clone http://127.0.0.1:3000/gitea/global-model.git
    pushd global-model

    git commit -m "pretrained model" --allow-empty
    git push origin master

    popd
    rm -rf global-model
    ```
4. Apply mnist configs
    ```bash
    kubectl apply -f configs.yml
    ```

5. Apply mnist deployment
    ```bash
    kubectl apply -f mnist-deployment.yml
    ```

6. Trigger training by updating train plan `plan.json` and pushing to the `train-plan` repository
    ```bash
    # Set port-forward to access gitea by `localhost`
    kubectl port-forward service/harmonia-gitea 3000
    ```
    ```bash
    git clone http://127.0.0.1:3000/gitea/train-plan.git

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

7. Shows the tensorboard view
    ```bash
    # Set port-forward to access gitea by `localhost`
    kubectl port-forward service/logserver 6006
    ```

    using browser surfing to `http://localhost:6006`
