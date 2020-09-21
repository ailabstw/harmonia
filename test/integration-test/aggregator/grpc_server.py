from concurrent import futures
import logging
import time
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
RETURNS = [
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
    {
        "metadata": {
            "a011": "a011_value",
            "a012": "a012_value",
        },
        "metrics": {
            "a011": 2.11,
            "a012": 2.12,
        },
        "weights": "01",
    },
]

def req_dict_compare(obj, expected_dict):
    for key in expected_dict.keys():
        if not hasattr(obj, key) or getattr(obj, key) != expected_dict[key]:
            return False
    return True

class AggregateServerApp(service_pb2_grpc.AggregateServerAppServicer):
    def __init__(self):
        self.round = 0

    def next_iteration(self):
        self.round += 1

    def request_validate(self, request):
        expected = EXPECTEDS[self.round]
        assert len(request.localModels) == 2, f"Number of localModels error. Expect [2] Actual [{len(request.localModels)}]"
        assert request.localModels[0].datasetSize == expected["datasetSize"], f"datasetSize error. [{expected['datasetSize']}] expected, [{request.localModels[0].datasetSize}] got"
        assert operator.eq(request.localModels[0].metadata, expected["metadata"]), f"Check metadata fail. [{expected['metadata']}] expected, [{request.localModels[0].metadata}] got"
        assert operator.eq(request.localModels[0].metrics, expected["metrics"]), f"Check metrics fail. [{expected['metrics']}] expected, [{request.localModels[0].metrics}] got"

        try:
            with open(os.path.join("/repos", request.localModels[0].path, "weights"), "r") as finput:
                weights = str(finput.read())
                assert weights == expected["weights"], f"Check 'weights' fail. [{expected['weights']}] expected, [{weights}] got"
        except Exception as exception:
            raise f"Read 'weights' fail. [{exception}]"

    def send_response(self, repo_path):
        LOGGER.info("starting aggregating")

        return_values = RETURNS[self.round]
        # to decrease the possibility of race condition
        time.sleep(5)
        with open(os.path.join("/repos", repo_path, "weights"), "w") as foutput:
            foutput.write(return_values["weights"])

        msg = service_pb2.AggregateResult(
            error=0,
            metadata=return_values["metadata"],
            metrics=return_values["metrics"],
        )

        try:
            channel = grpc.insecure_channel(GRPC_CLIENT_URI)
            stub = service_pb2_grpc.AggregateServerOperatorStub(channel)
            LOGGER.info("sending: [%s]", msg)
            response = stub.AggregateFinish(msg)
            LOGGER.info("response: [%s]", response)
        except grpc.RpcError as rpc_error:
            LOGGER.error(
                "grpc client call AggregateFinish failure: %s", rpc_error
            )

        self.next_iteration()

    def Aggregate(self, request, _):
        LOGGER.info("Aggregate: request [%s]", request)

        self.request_validate(request)

        threading.Thread(
            target=self.send_response,
            daemon=True,
            args=(request.aggregatedModel.path, )
        ).start()

        return service_pb2.Empty()

    def TrainFinish(self, _0, _1):
        LOGGER.info("Train Finish")
        STOP_EVENT.set()
        return service_pb2.Empty()

def serve():
    LOGGER.info("Start server...")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_AggregateServerAppServicer_to_server(
        AggregateServerApp(),
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
