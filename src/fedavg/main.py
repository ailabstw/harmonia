from concurrent import futures
import logging
import threading
import os
import grpc
import merge
import service_pb2
import service_pb2_grpc

OPERATOR_URI = os.getenv('OPERATOR_URI', '127.0.0.1:8787')
APPLICATION_URI = os.getenv('APPLICATION_URI', '0.0.0.0:7878')
LOG_LEVEL = os.environ.get('LOG_LEVEL', 'INFO')
REPO_ROOT = os.environ.get('REPO_ROOT', '/repos')
MODEL_FILENAME = os.environ.get('MODEL_FILENAME', 'weights.tar')
STOP_EVENT = threading.Event()

AGGREGATE_SUCCESS = 0
AGGREGATE_CONDITION = 1
AGGREGATE_FAIL = 2

logging.basicConfig(level=LOG_LEVEL)

def send_result(err):
    logging.info("config.GRPC_CLIENT_URI: [%s]", OPERATOR_URI)
    try:
        channel = grpc.insecure_channel(OPERATOR_URI)
        stub = service_pb2_grpc.AggregateServerOperatorStub(channel)
        res = service_pb2.AggregateResult(error=err,)
        response = stub.AggregateFinish(res)
    except grpc.RpcError as rpc_error:
        logging.error("grpc error: [%s]", rpc_error)
    except Exception as err:
        logging.error("got error: [%s]", err)

    logging.debug("sending grpc message succeeds, response: [%s]", response)

def aggregate(local_models, aggregated_model):
    if len(local_models) == 0:
        send_result(AGGREGATE_FAIL)
        return

    models = []
    for local_model in local_models:
        path = os.path.join(REPO_ROOT, local_model.path, MODEL_FILENAME)
        if os.path.isfile(path):
            models.append({'path': path, 'size': local_model.datasetSize})
    output_path = os.path.join(REPO_ROOT, aggregated_model.path, MODEL_FILENAME)

    logging.debug("models: [%s]", models)
    logging.debug("output_path: [%s]", output_path)
    merge.merge(models, output_path)

    send_result(AGGREGATE_SUCCESS)

class AggregateServerServicer(service_pb2_grpc.AggregateServerAppServicer):
    def Aggregate(self, request, context):
        logging.info("received Aggregate message [%s]", request)

        threading.Thread(
            target=aggregate,
            args=(request.localModels, request.aggregatedModel),
            daemon=True
        ).start()

        response = service_pb2.Empty()
        return response

    def TrainFinish(self, _request, _context):
        logging.info("received TrainFinish message")
        STOP_EVENT.set()
        return service_pb2.Empty()

def serve():
    logging.info("Start server... [%s]", APPLICATION_URI)

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_AggregateServerAppServicer_to_server(
        AggregateServerServicer(), server)
    server.add_insecure_port(APPLICATION_URI)
    server.start()

    STOP_EVENT.wait()
    logging.info("Server Stop")
    server.stop(None)

if __name__ == "__main__":
    serve()
