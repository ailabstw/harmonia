import grpc_client
from logger import setup_logger
import os
import time


logger = setup_logger(__name__)


def start_training(repo_path):
    logger.info("starting training")
    # to decrease the possibility of race condition
    time.sleep(5)
    with open(os.path.join("/repos", repo_path, "README.md"), 'a') as f:
        f.write("new line")
    logger.info("training succeed")
    grpc_client.signalModelTrainingCompleted()


if __name__ == "__main__":
    start_training(".")
