from concurrent import futures
import logging
import threading
import os
import random
import grpc
import service_pb2
import service_pb2_grpc
import mnist

OPERATOR_URI = os.getenv('OPERATOR_URI', "localhost:8787")
APPLICATION_URI = "0.0.0.0:7878"
__DATA = []
STOP_EVENT = threading.Event()

def get_training_data():
    global __DATA
    if not __DATA:
        __DATA = random.sample(range(60000), 2000)
    return __DATA

def train(baseModel, output_model_path, epochs=1):
    data = get_training_data()
    output = os.path.join("/repos", output_model_path, 'weights.tar')
    logging.info(f'input path: [{baseModel.path}]')
    logging.info(f'output path: [{output}]')
    logging.info(f'epochs: {epochs}')

    base_weight_path = os.path.join("/repos", baseModel.path, "weights.tar")
    try:
        metrics = mnist.train(data, output, epochs=epochs, resume=base_weight_path)
    except Exception as err:
        print(err)

    # Send finish message
    logging.info(f"GRPC_CLIENT_URI: {OPERATOR_URI}")
    try:
        channel = grpc.insecure_channel(OPERATOR_URI)
        stub = service_pb2_grpc.EdgeOperatorStub(channel)
        result = service_pb2.LocalTrainResult(
            error=0,
            datasetSize=2500,
            metrics=metrics
        )

        response = stub.LocalTrainFinish(result)
    except grpc.RpcError as rpc_error:
        logging.error("grpc error: {}".format(rpc_error))
    except Exception as err:
        logging.error('got error: {}'.format(err))

    logging.debug("sending grpc message succeeds, response: {}".format(response))


class EdgeAppServicer(service_pb2_grpc.EdgeAppServicer):
    def TrainInit(self, request, context):
        logging.info("TrainInit")
        resp = service_pb2.Empty()
        logging.info(f"Sending response: {resp}")
        return resp

    def LocalTrain(self, request, context):
        logging.info("LocalTrain")

        threading.Thread(
            target=train,
            args=(request.baseModel, request.localModel.path, request.EpR),
            daemon=True
        ).start()

        resp = service_pb2.Empty()
        logging.info("Sending response: {}".format(resp))
        return resp

    def TrainInterrupt(self, _request, _context):
        # Not Implemented
        return service_pb2.Empty()

    def TrainFinish(self, _request, _context):
        logging.info("TrainFinish")
        STOP_EVENT.set()
        return service_pb2.Empty()


def serve():
    logging.basicConfig(level=logging.DEBUG)

    logging.info("Start server... {}".format(APPLICATION_URI))

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_EdgeAppServicer_to_server(
        EdgeAppServicer(), server)
    server.add_insecure_port(APPLICATION_URI)
    server.start()

    STOP_EVENT.wait()
    logging.info("Server Stop")
    server.stop(None)

if __name__ == "__main__":
    serve()
