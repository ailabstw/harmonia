import yaml

GRPC_SERVER_URI = ":7878"
GRPC_CLIENT_URI = ":8787"
with open("config.yml") as f:
    docs = yaml.load(f, Loader=yaml.FullLoader)
    GRPC_SERVER_URI = docs['appGrpcServerURI']
    GRPC_CLIENT_URI = docs['operatorGrpcServerURI']
