# MNIST Example

Here shows how to build grpc packages and `Application` images.

# Generate gRPC Python Modules
`Application` relies on Harmonia defined gRPC protocol (`service.proto`) to interact with `Operator`. Following shows how to generate two modules (`*_pb2.py`, `*_grpc_pb2.py`) that should be imported by python implemented `Application`.

```bash
make -C ../src/protos python_protos
```

# Build Application Images
Copy `service_pb2.py`, `service_pb2_grpc.py` into `edge` to build the image.

```bash
cp -pv ../src/protos/python_protos/* edge
docker build -t mnist_edge edge
```

# More
Check subfolders for different platform deployments:
*  [docker](./docker_deployment)
*  [kubernetes](./k8s_deployment)
