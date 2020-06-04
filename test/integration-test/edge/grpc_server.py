from concurrent import futures
import grpc
from threading import Thread

from logger import setup_logger
import app
import config
import service_pb2
import service_pb2_grpc


logger = setup_logger(__name__)


class EdgeAppServicer(service_pb2_grpc.EdgeApp):

    def LocalTrain(self, request, context):
        t = Thread(target=app.start_training, daemon=True, args=(request.outputModelPath, ))
        t.start()
        msg = service_pb2.Msg(message="ok")
        logger.info("Sending response: {}".format(msg))
        return msg


def serve():
    logger.info("Start gRpc server binding on {}...".format(
        config.GRPC_SERVER_URI))
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_EdgeAppServicer_to_server(EdgeAppServicer(), server)
    logger.info('listen on [::]:' + config.GRPC_SERVER_URI.split(':')[1])
    server.add_insecure_port('[::]:' + config.GRPC_SERVER_URI.split(':')[1])
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
