from concurrent import futures
import grpc
import logging
import time
import service_pb2
import service_pb2_grpc
from threading import Thread
import os
import merge

OPERATOR_URI = "127.0.0.1:8787"
APPLICATION_URI = "0.0.0.0:7878"

def aggregate(input_paths, output_path):
    input_full_paths = [os.path.join("/repos", d, 'weights.tar') for d in input_paths if os.path.isfile(
        os.path.join("/repos", d, 'weights.tar'))]
    output_full_path = os.path.join("/repos", output_path, "merged.tar")

    try:
        merge.merge(input_full_paths, output_full_path)
    except Exception as err:
        logging.error("grpc error: {}".format(err))

    # Send finish message
    logging.info("config.GRPC_CLIENT_URI: {}".format(OPERATOR_URI))
    try:
        channel = grpc.insecure_channel(OPERATOR_URI)
        stub = service_pb2_grpc.AggregateServerOperatorStub(channel)
        msg = service_pb2.Msg(message="finish")

        response = stub.AggregateFinish(msg)
    except grpc.RpcError as rpc_error:
        logging.error("grpc error: {}".format(rpc_error))
    except Exception as err:
        logging.error('got error: {}'.format(err))

    logging.debug("sending grpc message succeeds, response: {}".format(response))

class AggregateServerServicer(service_pb2_grpc.AggregateServerAppServicer):
    def Aggregate(self, request, context):
        logging.debug("Aggregate")

        # Use another thread do training
        t = Thread(target=aggregate,
                   args=(request.inputModelPaths, request.outputModelPath),
                   daemon=True)
        t.start()
        msg = service_pb2.Msg(message="ok")
        logging.info("Sending response: {}".format(msg))
        return msg


def serve():
    logging.basicConfig(level=logging.DEBUG)

    logging.info("Start server... {}".format(APPLICATION_URI))

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_AggregateServerAppServicer_to_server(
        AggregateServerServicer(), server)
    server.add_insecure_port(APPLICATION_URI)
    server.start()
    while True:
        time.sleep(10)


if __name__ == "__main__":
    serve()
