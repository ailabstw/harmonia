import subprocess
from logger import setup_logger
import config
import grpc_client
import os
import time


logger = setup_logger(__name__)


def start_aggregating(repo_path):
    logger.info("starting aggregating")
    # to decrease the possibility of race condition
    time.sleep(5)
    with open(os.path.join("/repos", repo_path, "README.md"), 'a') as f:
        f.write('new line')
    logger.info("aggregating succeed")
    grpc_client.signalAggregatorCompleted()


if __name__ == "__main__":
    start_aggregating()
