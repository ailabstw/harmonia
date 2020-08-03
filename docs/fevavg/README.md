# FedAvg Image Usage
`harmonia/fedavg` is the pytorch implementation of FedAvg in Harmonia.

## Build
```bash
pushd src/fedavg
make
```

## Configuration

|Name             |Description                              |Required|
|---              |---                                      |---     |
|OPERATOR_URI     | URI that `harmonia/avg` sends message to| Optional: default `127.0.0.1:8787` |
|APPLICATION_URI  | URI that the gRPC server binds          | Optional: default `0.0.0.0:7878`  |
|LOG_LEVEL        | One of `ERROR`, `INFO`, `DEBUG`         | Optional: default `INFO` |
|REPO_ROOT        | Mount point of the shared volume        | Optional: default `/repos` |
|MODEL_FILENAME   | Filename that `harmonia/avg` reads and writes models to | Optional: default `weights.tar` |


## Docker Compose Usage
This only shows `docker-compose.yml` of the aggregator part:  

```yml
version: "3.7"
services:
  app:
    image: harmonia/fedavg
    environment:
      OPERATOR_URI: operator:8787
    volumes:
      - type: volume
        source: shared
        target: /repos
  operator:
    image: harmonia/operator
    volumes:
      - ./config.yml:/app/config.yml
      - type: volume
        source: shared
        target: /repos
volumes:
  shared:
```

## Edge Implementation Guild
With default configurations, `harmonia/fedavg` reads models by
```python
# '/repos/{model_relative_path}/weights.tar'
torch.load(os.path.join(REPO_ROOT, model_relative_path, MODEL_FILENAME))
```

Correspondingly, edge applications should save models by
```python
torch.save(model.state_dict(), f'/repos/{model_relative_path}/weights.tar')
```
