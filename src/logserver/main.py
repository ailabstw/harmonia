import logging
from .config import cfg
from multiprocessing import Process
from . import webhook
from .webhook import Webhook, process_webhook
from . import gitoperation
from .gitoperation import ModelRepo
from flask import Flask
from flask import request
from flask import Response


logger = logging.getLogger(__name__)
app = Flask(__name__)


@app.route('/', methods=['POST'])
def get_webhook():
    logger.info('received webhook')
    webhook = Webhook(request.json['repository']['full_name'], request.json['ref'])
    p = Process(target=process_webhook, args=(webhook,))
    p.start()
    logger.debug(f'webhook full name: {webhook.repo.full_name}')

    response = Response()
    response.status_code = 200
    return response


def setup_repos(repos: 'ModelRepo'):
    for repo in repos:
        repo_path = gitoperation.get_repo_path(repo.git_http_url)
        gitoperation.clone_repo(repo_path, repo.git_http_url)
        gitoperation.setup_git(repo_path,
                               gitoperation.GitUser(
                                   "Harmonia Operator",
                                   "operator@harmonia",
                                   cfg.git_user_token))


if __name__ == "__main__":
    setup_repos(cfg.model_repos)
    host, port = cfg.steward_server_URI.split(':')
    app.run(host=host, port=port)
