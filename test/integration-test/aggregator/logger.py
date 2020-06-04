import logging


def setup_logger(name, level=logging.INFO):
    logging.basicConfig(
        format='%(asctime)s %(levelname)s %(message)s', level=level)
    logger = logging.getLogger(name)
    return logger
