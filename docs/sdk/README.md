# Harmonia Operator SDK
This README shows usage of `harmonia/operator` image. contents:

- [System Architecture](#system-architecture)
- [System Components](#system-compoments)
    - [Gitea](#giteahttpsgiteaioen-us)
        - [Repositories](#repositories)
    - [Aggregator/Edge Nodes](#aggregatoredge-nodes)
        - [Shared Storage between Operator and Application](#shared-storage)
        - [Operator Container](#operator-container)
            - [Configuration](#configuration)
        - [Application Container](#application-container)
- [FL System States & gRPC protocols](#fl-system-states-grpc--protocols)
    - [Aggregator](#aggregator)
    - [Edge](#edge)

---

## System Architecture
<div align="center"><img src="../../assets/detail_architecture.jpg" width="50%" height="50%"></div>

The above figure illustrates the Harmonia configuration with two local training nodes. The numbers shown in the figure indicate the steps to finish a federated learning cycle in the release. To start a FL training, a training plan is registered in the git registry (1), and the registry notifies all the participants via webhook (2). Two local nodes are then triggered to start local training (3). Once a local node completes its local training, the resulting model (called local model) is pushed to the registry (4) and the aggregator pulls this local model (5). Once the aggregator receives local models of all the participants, it performs model aggregation (6), and the aggregated model is also pushed to the git registry (7). The aggregated model is then pulled to local nodes to start another round of local training (8). This process is repeated until a user-defined converge condition is met.

---

## System Compoments
### [Gitea](https://gitea.io/en-us/)
Gitea is an open source git registry. Changes to repositories, e.g., models updates, trigger FL system state transitions via webhooks. 

#### Repositories
We have three types of repositories in Gitea:
1. Training Plan: it stores the required parameters for a FL cycle. The training plan should be a json file named `plan.json` :
    ```json
    {
        "roundCount": 100,
        "edgeCount": 2,
        "epochCount": 1
    }
    ```
    |Field      |Description                                |
    |---        |---                                        |
    |roundCount | number of rounds in this FL training job  |
    |edgeCount  | number of edges                           |
    |epochCount | number of epochs per round                |
2. Aggregated Model: it stores aggregated models pushed by the Aggregator container. The final aggregated model is tagged with `inference-<commit_hash_of_train_plan>`.
3. Edge Models: these repositories store local models pushed by each nodes seperatedly.

### Aggregator/Edge Nodes
An FL participant is activated as a K8S Pod, which contains an `Operator` container and `Application` container. 

===

#### Shared Storage
This volume stores git cloned repositories and is shared `operator` and `application`.

===

#### Operator Container
Harmonia operator, an instance of `harmonia/operator` image,  is responsible for:
1. handling Gitea webhooks and pulling updates from Gitea.
2. serving `gRPC` requests from an `Application` container 
3. managing FL states, and taking corresponding actions, e.g., pushing models to Gitea or sending messages to an `Application` container.

##### Configuration
`Operator` can be configured with `/app/config.yml`. Each field in the configuration file is described below:
```yaml
type: aggregator
gitUserToken: <aggregator_token>
aggregatorModelRepo:
    gitHttpURL: http://<aggregator_account>@<gitea_URL>/<global_model_repo>.git
edgeModelRepos:
    - gitHttpURL: http://<aggregator_account>@<gitea_URL>/<local_model1_repo>.git
    - gitHttpURL: http://<aggregator_account>@<gitea_URL>/<local_model2_repo>.git
trainPlanRepo:
    gitHttpURL: http://<aggregator_account>@<gitea_URL>/<train_plan_repo>.git
```

|Field      |Description|Required   |
|---        |---        |---        |
|type| one of `aggregator` or `edge`|Required|
|logLevel| one of `debug`, `info`, `warn`, `error`, `panic`, `fatal`|Optional: default `info`|
|logPath| absolute path to log |Optional: default `""`, console log only|
|stewardServerURI| listens gitea webhooks on `hostname:port` |Optional: default `0.0.0.0:9080`|
|operatorGrpcServerURI| listens application gRPC message on `hostname:port` |Optional: default `localhost:8787`|
|appGrpcServerURI| location that application gRPC server listens to.  |Optional: default `localhost:7878`|
|gitUserToken| the token of Gitea user |Required|
|aggregatorModelRepo| A `RepositoryDesp` of aggregated model repository |Required|
|trainPlanRepo| A `RepositoryDesp` of train plan repository |Required|
|edgeModelRepos| An `RepositoryDesp` array of edge model repositories \\ |Required by aggregator operator|
|edgeModelRepo.gitHttpURL| A `RepositoryDesp` edge model repository  |Required by edge operator|

|Type       |Description|
|---        |---        |
|RepositoryDesp|An repository medata contains `gitHttpURL`|
|gitHttpURL|format: `http://<username>@<gitea_URL>/<repo>.git`|

===

#### Application Container
Local training or aggregation applications is encapsulated in an `Application` container, which is implemented by users. An `Application` container is responsible for

1. serving `gRPC` requests from operator 
1. performing local model training or model aggregation 
1. sending back process status to operator through Harmonia operator protocol descrobed below  
​  
Users should build gRPC services via `protos/service.proto` with their prefered language and message handlers should be implemented . See [gRPC tutorial](https://grpc.io/docs/tutorials/).  
Following is a piece of aggregator `Application` for demonstration: (see [mnist example](../../examples/mnist) for full example)
```python
import grpc
import os
import service_pb2
import service_pb2_grpc

class AggregateServerServicer(service_pb2_grpc.AggregateServerAppServicer):
    def Aggregate(self, request, context):
        f = os.fork()
        if pid > 0:
            msg = service_pb2.Msg(message="ok")
            return msg
        else:
            # Training Code
            channel = grpc.insecure_channel(OPERATOR_URI)
            stub = service_pb2_grpc.AggregateServerOperatorStub(channel)
            msg = service_pb2.Msg(message="finish")

            response = stub.AggregateFinish(msg)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_AggregateServerAppServicer_to_server(AggregateServerServicer(), server)
    server.add_insecure_port(APPLICATION_URI)
    server.start()
    while True:
        time.sleep(10)

if __name__ == "__main__":
    serve()
```

---

### FL System States & gRPC protocols
#### Aggregator
<div align="center"><img src="../../assets/aggregator_state_diagram.jpg"  width="50%" height="50%"></div>

gRPC protocol:
* Message to `application`
    * Aggregate:  
    This event is sent when there are `edgeCount` edges finished local train and models are stored in shared volume ready for merging.  
        **Inputs:**

        |Name           |Type       |Description    |
        |---            |---        |---            |
        |inputModelPaths| string[]  | relative paths of edge models from shared storage|
        |outputModelPath| string    | relative paths of aggregated model from shared storage|

        **outputs**
        (None)

* Message from `application`
    * AggregateFinish:  
    This action is sent without payload after user finishes merge and places the `aggregated model` into `outputModelPath` specified by `Aggregate` event.
        **Inputs:**
        (None)
        **outputs**
        (None)

#### Edge
<div align="center"><img src="../../assets/edge_state_diagram.jpg"  width="50%" height="50%"></div>

gRPC protocol:
* Message to `application`
    * LocalTrain:
        This event is sent when receiving an `aggregated model` indicating another local train is ready to process.  
        **Inputs:**  

        |Name           |Type       |Description    |
        |---            |---        |---            |
        |inputModelPath | string    | relative path of current base model from shared storage|
        |outputModelPath| string    | relative path of local model from shared storage|
        |epochCount     | int32     | number of epochs per round|

        **Outputs:**
        (None)

* Message from `application`
    * LocalTrainFinish:
    This action is sent without payload after user finishes local train.
        **Inputs:**
        (None)
        **Outputs:**
        (None)
