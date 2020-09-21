from concurrent import futures
import logging
import os
import operator
import threading
import grpc
import service_pb2
import service_pb2_grpc

STOP_EVENT = threading.Event()

GRPC_SERVER_URI = "0.0.0.0:7878"
GRPC_CLIENT_URI = "operator:8787"

logging.basicConfig(
    format="%(asctime)s %(levelname)s %(message)s",
    level=logging.DEBUG,
)
LOGGER = logging.getLogger(__name__)

TRAIN_PLAN = {
    "roundCount": 2,
    "EpR": 3,
}
EXPECTEDS = [
    {
        "metadata": {},
        "metrics": {},
        "weights": "",
    },
    {
        "metadata": {
            "a001": "a001_value",
            "a002": "a002_value",
        },
        "metrics": {
            "a001": 2.01,
            "a002": 2.02,
        },
        "weights": "00",
    },
]
RETURNS = [
    {
        "datasetSize": 100,
        "metadata": {
            "m001": "m001_value",
            "m002": "m002_value",
        },
        "metrics": {
            "m001": 0.01,
            "m002": 0.02,
        },
        "weights": "00",
    },
    {
        "datasetSize": 200,
        "metadata": {
            "m011": "m011_value",
            "m012": "m012_value",
        },
        "metrics": {
            "m011": 0.11,
            "m012": 0.12,
        },
        "weights": "01",
    },
]

class EdgeAppServicer(service_pb2_grpc.EdgeAppServicer):
    def __init__(self):
        self.round = 0

    def next_iteration(self):
        self.round += 1

    def request_validate(self, request):
        expected = EXPECTEDS[self.round]
        assert operator.eq(request.baseModel.metadata, expected["metadata"]), f"Check metadata fail. [{expected['metadata']}] expected, [{request.baseModel.metadata}] got"
        assert operator.eq(request.baseModel.metrics, expected["metrics"]), f"Check metrics fail. [{expected['metrics']}] expected, [{request.baseModel.metrics}] got"

        if expected["weights"] == "":
            return

        try:
            with open(os.path.join("/repos", request.baseModel.path, "weights"), "r") as finput:
                weights = str(finput.read())
                assert weights == expected["weights"], f"Check 'weights' fail. [{expected['weights']}] expected, [{weights}] got"
        except Exception as exception:
            raise f"Read 'weights' fail. [{exception}]"

    def send_response(self, repo_path):
        LOGGER.info("starting local train")

        return_values = RETURNS[self.round]

        with open(os.path.join("/repos", repo_path, "weights"), "w") as foutput:
            foutput.write(return_values["weights"])

        msg = service_pb2.LocalTrainResult(
            error=0,
            datasetSize=return_values["datasetSize"],
            metadata=return_values["metadata"],
            metrics=return_values["metrics"],
        )

        try:
            channel = grpc.insecure_channel(GRPC_CLIENT_URI)
            stub = service_pb2_grpc.EdgeOperatorStub(channel)
            LOGGER.info("sending: [%s]", msg)
            response = stub.LocalTrainFinish(msg)
            LOGGER.info("response: [%s]", response)
        except grpc.RpcError as rpc_error:
            LOGGER.error(
                "grpc client call LocalTrainFinish failure: %s", rpc_error
            )

        self.next_iteration()

    def TrainInit(self, _0, _1):
        LOGGER.info("Initialize train")
        return service_pb2.Empty()

    def LocalTrain(self, request, _):
        LOGGER.info("LocalTrain: request [%s]", request)

        self.request_validate(request)

        threading.Thread(
            target=self.send_response,
            daemon=True,
            args=(request.localModel.path, )
        ).start()

        return service_pb2.Empty()

    def TrainFinish(self, _0, _1):
        LOGGER.info("Train Finish")
        STOP_EVENT.set()
        return service_pb2.Empty()

def serve():
    LOGGER.info("Start server...")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_EdgeAppServicer_to_server(
        EdgeAppServicer(),
        server,
    )

    LOGGER.info("listen on [%s]", GRPC_SERVER_URI)
    server.add_insecure_port("[::]:" + GRPC_SERVER_URI.split(":")[1])
    server.start()

    STOP_EVENT.wait()
    LOGGER.info("Server Stop")
    server.stop(None)


if __name__ == "__main__":
    serve()
