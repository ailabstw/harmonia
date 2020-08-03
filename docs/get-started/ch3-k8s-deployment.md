# Deploy Harmonia FL
This chapter shows all images you should prepare for a Harmonia FL and how to deploy them on k8s.

## Build Harmonia Operator
Simply install `harmonia/operator` image by:
```bash
make all
```

## Build Applications
1. Build python grpc packages in `src/protos`
    ```bash
    cd src/protos
    make python_protos
    ```  
    And copy `service_pb2.py`, `service_pb2_grpc.py` into `examples/edge`.


2. Build `application` image
    ```bash
    docker build examples/edge --tag <image_registry>/mnist-edge
    ````

## Deploy on K8S
1. Push images to image registry
    ```bash
    docker tag operator <image_registry>/harmonia/operator
    docker push <image_registry>/harmonia/operator
    docker push <image_registry>/mnist-edge
    ```

3. Create k8s configuration
    ```yml
    # mnist-deployment.yml

    # Aggregator
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: aggregator
      labels:
        app: aggregator
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: aggregator
      template:
        metadata:
          labels:
            app: aggregator
        spec:
          # Each Harmonia participant are composed by an operator container (provided by Harmonia)
          # and an user application container
          containers:
          # operator container
          - name: operator
            image: <image_registry>/harmonia/operator
            imagePullPolicy: Always
            ports:
            - containerPort: 9080
              name: steward
            volumeMounts:
            # Operator configuration is provided in chapter 2
            - name: config
              mountPath: /app/config.yml
              subPath: aggregator-config.yml
            # Shared storage with application container
            - name: shared-repos
              mountPath: /repos
          # application container
          - name: application
            image: <image_registry>/harmonia/fedavg
            imagePullPolicy: Always
            volumeMounts:
            # Shared storage with operator container
            - name: shared-repos
              mountPath: /repos
          volumes:
          - name: config
            configMap:
              name: aggregator-config
          - name: shared-repos
            emptyDir: {}

    ---
    # Service is used to expose k8s pod to Gitea and listen webhook
    # With default operator configuration, 9080 port is exposed.
    # If Gitea service is deployed in the same `mnist-aggregator:9080`
    kind: Service
    apiVersion: v1
    metadata:
      name: mnist-aggregator
    spec:
      selector:
        app: aggregator
      ports:
      - name: aggregator
        port: 9080
        targetPort: 9080
      type: NodePort

    ---
    # Edge1
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: edge1
      labels:
        app: edge1
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: edge1
      template:
        metadata:
          labels:
            app: edge1
        spec:
          containers:
          - name: operator
            image: <image_registry>/harmonia/operator
            imagePullPolicy: Always
            ports:
            - containerPort: 9080
              name: steward
            volumeMounts:
            - name: config
              mountPath: /app/config.yml
              subPath: edge-config.yml
            - name: shared-repos
              mountPath: /repos
          - name: application
            image: <image_registry>/mnist-edge
            imagePullPolicy: Always
            volumeMounts:
            - name: shared-repos
              mountPath: /repos
          volumes:
          - name: config
            configMap:
              name: edge1-config
          - name: shared-repos
            emptyDir: {}

    ---
    kind: Service
    apiVersion: v1
    metadata:
      name: mnist-edge1
    spec:
      selector:
        app: edge1
      ports:
      - name: edge1
        port: 9080
        targetPort: 9080
      type: NodePort

    ---
    # Edge2
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: edge2
      labels:
        app: edge2
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: edge2
      template:
        metadata:
          labels:
            app: edge2
        spec:
          containers:
          - name: operator
            image: <image_registry>/harmonia/operator
            imagePullPolicy: Always
            ports:
            - containerPort: 9080
              name: steward
            volumeMounts:
            - name: config
              mountPath: /app/config.yml
              subPath: edge-config.yml
            - name: shared-repos
              mountPath: /repos
          - name: application
            image: <image_registry>/mnist-edge
            imagePullPolicy: Always
            volumeMounts:
            - name: shared-repos
              mountPath: /repos
          volumes:
          - name: config
            configMap:
              name: edge2-config
          - name: shared-repos
            emptyDir: {}

    ---
    kind: Service
    apiVersion: v1
    metadata:
      name: mnist-edge2
    spec:
      selector:
        app: edge2
      ports:
      - name: edge2
        port: 9080
        targetPort: 9080
      type: NodePort
    ```

4. Apply to k8s
    ```bash
    kubectl apply -f mnist-deployment.yml
    ```

5. Investigate the pod states
    ```bash
    kubectl get pod
    ```

While the `STATUS` of the pods `aggregator-xxxx`,`edge1-xxxx` and `edge2-xxxx` all become `Running`, Harmonia FL is ready and [next chapter](ch4-execute-learning.md) would is the last step to start learning processes.
