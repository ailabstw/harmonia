import logging
from .config import cfg
from . import gitoperation
from . import metadata


logger = logging.getLogger(__name__)


class Webhook:
    class Repo:
        def __init__(self, full_name):
            self.full_name = full_name

    def __init__(self, repo_name, ref):
        self.repo = self.Repo(repo_name)
        tokens = ref.split('/')
        if tokens[1] == 'heads':
            self.ref = 'origin/' + tokens[2]
        elif tokens[1] == 'tags':
            self.ref = tokens[2]
        else:
            logger.error(f'unknown webhook reference type: {tokens[1]}')


def process_webhook(webhook: 'Webhook'):
    logger.info('processing webhook...')
    for model_repo in cfg.model_repos:
        if gitoperation.convert_git_http_url_to_full_name(model_repo.git_http_url) == webhook.repo.full_name:
            logger.info(f'The webhook sent from model_repo git http url {model_repo.git_http_url}')
            repo_path = gitoperation.get_repo_path(model_repo.git_http_url)
            m = metadata.get_metadata(webhook.repo.full_name, webhook.ref, repo_path)
            if m:
                metadata.add_record_to_tensorboard(m)
            break
