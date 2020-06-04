import grpc

from logger import setup_logger
import config
import service_pb2
import service_pb2_grpc


logger = setup_logger(__name__)


def signalAggregatorCompleted():
    logger.info("config.GRPC_CLIENT_URI: {}".format(config.GRPC_CLIENT_URI))
    try:
        channel = grpc.insecure_channel(config.GRPC_CLIENT_URI)
        stub = service_pb2_grpc.AggregateServerOperatorStub(channel)
        msg = service_pb2.Msg(message="finish")
        try:
            logger.info("sending: {}".format(msg))
            response = stub.AggregateFinish(msg)
            logger.info("response: {}".format(response))
        except grpc.RpcError as rpc_error:
            logger.error(
                "grpc client call AggregateFinish failure: %s", rpc_error)
        logger.info('finish work')
    except Exception as e:
        logger.error('got error: {}'.format(e))
