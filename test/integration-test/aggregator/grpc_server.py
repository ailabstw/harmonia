from concurrent import futures
import grpc
import threading
from threading import Thread

from logger import setup_logger
import app
import config
import service_pb2
import service_pb2_grpc


logger = setup_logger(__name__)


class AggregateServerApp(service_pb2_grpc.AggregateServerApp):

    def Aggregate(self, request, context):
        t = Thread(target=app.start_aggregating, daemon=True, args=(request.outputModelPath, ))
        t.start()
        msg = service_pb2.Msg(message="ok")
        logger.info("Sending response: {}".format(msg))
        return msg


def serve():
    logger.info("Start server...")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_AggregateServerAppServicer_to_server(
        AggregateServerApp(), server)
    logger.info('listen on [::]:' + config.GRPC_SERVER_URI.split(':')[1])
    server.add_insecure_port('[::]:' + config.GRPC_SERVER_URI.split(':')[1])
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
