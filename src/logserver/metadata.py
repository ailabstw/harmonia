from . import gitoperation
import json
import logging
import os
import tensorflow as tf
from .config import cfg


logger = logging.getLogger(__name__)
FILENAME = os.getenv('METADATA_FILENAME', '.harmonia')


class Metadata:
    def __init__(self, node_name, plan_id, round_number, dataset_size, metrics):
        self.dataset_size = dataset_size
        self.round_number = round_number
        self.plan_id = plan_id
        self.metrics = metrics
        self.node_name = node_name


def get_metadata(repo_name: str, ref:str, repo_path: str):
    logger.info('getting metadata content...')
    gitoperation.fetch_repo(repo_path)

    gitoperation.checkout_file(repo_path, ref, FILENAME)

    filepath = os.path.join(repo_path, FILENAME)
    if not os.path.exists(filepath):
        logger.warning(f'{FILENAME} not exists')
        return None

    with open(filepath, 'r') as f:
        data = json.load(f)
        logger.debug(f'content of metadata: {data}')
        dataset_size = data['datasetSize'] if 'datasetSize' in data else None
        round_number = data['roundNumber'] if 'roundNumber' in data else None
        plan_id = data['trainPlanID'] if 'trainPlanID' in data else None
        metrics = data['metrics'] if 'metrics' in data else None
        metadata = Metadata(repo_name, plan_id,
                            round_number, dataset_size, metrics)
        logger.info(
            f"""
            the Metadata instance has the values:
            node_name={metadata.node_name},
            plan_id={metadata.plan_id},
            round_number={metadata.round_number},
            dataset_size={metadata.dataset_size},
            metrics={metadata.metrics},
            """
        )
    return metadata


def add_record_to_tensorboard(metadata: 'Metadata'):
    logger.info('add record to tensorboard')
    writer = '{}/{}/{}'.format(cfg.tensorboard_data_root_dir,
                               metadata.node_name, metadata.plan_id)
    logger.info(f'current tensorbaord writer: {writer}')
    train_summary_writer = tf.summary.create_file_writer(writer)
    with train_summary_writer.as_default():
        metrics = metadata.metrics
        for k in metrics:
            if metrics[k] is None:
                logger.info(f'metrics[{k}] is None')
                continue
            logger.info(f'add record {k}: {metrics[k]}')
            tf.summary.scalar(k, metrics[k], step=metadata.round_number)
