from concurrent import futures
import grpc
import logging
import time
import service_pb2
import service_pb2_grpc
from threading import Thread
import os
import random
import mnist

OPERATOR_URI = "127.0.0.1:8787"
APPLICATION_URI = "0.0.0.0:7878"
__DATA = []

def get_training_data():
    global __DATA
    if not __DATA:
        __DATA = random.sample(range(60000), 2000)
    return __DATA

def train(input_model_path, output_model_path, epochs=1):
    data = get_training_data()
    output = os.path.join("/repos", output_model_path, 'weights.tar')
    logging.info('input path: [{}]'.format(input_model_path))
    logging.info('output path: [{}]'.format(output))
    logging.info('epochs: {}'.format(epochs))

    merged_weight_path = os.path.join("/repos", input_model_path, "merged.tar")
    try:
        mnist.train(data, output, epochs=epochs, resume=merged_weight_path)
    except Exception as err:
        print(err)

    # Send finish message
    logging.info("config.GRPC_CLIENT_URI: {}".format(OPERATOR_URI))
    try:
        channel = grpc.insecure_channel(OPERATOR_URI)
        stub = service_pb2_grpc.EdgeOperatorStub(channel)
        msg = service_pb2.Msg(message="finish")

        response = stub.LocalTrainFinish(msg)
    except grpc.RpcError as rpc_error:
        logging.error("grpc error: {}".format(rpc_error))
    except Exception as err:
        logging.error('got error: {}'.format(err))

    logging.debug("sending grpc message succeeds, response: {}".format(response))


class EdgeAppServicer(service_pb2_grpc.EdgeAppServicer):
    def LocalTrain(self, request, context):
        logging.info("LocalTrain")

        t = Thread(target=train,
                   args=(request.inputModelPath, request.outputModelPath, request.epochCount),
                   daemon=True)
        t.start()
        msg = service_pb2.Msg(message="ok")
        logging.info("Sending response: {}".format(msg))
        return msg


def serve():
    logging.basicConfig(level=logging.DEBUG)

    logging.info("Start server... {}".format(APPLICATION_URI))

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_EdgeAppServicer_to_server(
        EdgeAppServicer(), server)
    server.add_insecure_port(APPLICATION_URI)
    server.start()
    while True:
        time.sleep(10)


if __name__ == "__main__":
    serve()
