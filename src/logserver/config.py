import logging
import os
import yaml
from .gitoperation import ModelRepo
from threading import Lock

CONFIG_FILEPATH = os.getenv('CONFIG_FILEPATH', '/app/config.yml')

class Singleton(type):
    _instances = {}
    _lock: Lock = Lock()

    def __call__(cls, *args, **kwargs):
        """
        Possible changes to the value of the `__init__` argument do not affect
        the returned instance.
        """
        with cls._lock:
            if cls not in cls._instances:
                instance = super().__call__(*args, **kwargs)
                cls._instances[cls] = instance
        return cls._instances[cls]


class Config(metaclass=Singleton):

    def __init__(self, config_path=CONFIG_FILEPATH):
        self.steward_server_URI = "0.0.0.0:9080"
        self.tensorboard_data_root_dir = "/tensorboard_data"
        self.model_repos = []
        if not config_path is None:
            if os.path.exists(config_path):
                y = self.__read_yaml(config_path)
                if y:
                    self.__parse_yaml(y)
                else:
                    raise ValueError('empty yaml.')
            else:
                logging.warn(f'config file not found at {config_path}, using default settings')

    def __read_yaml(self, config_path: str) -> 'yaml':
        f = open(config_path, 'r')
        stream = f.read()
        y = yaml.load(stream, Loader=yaml.FullLoader)
        f.close()
        return y

    def __parse_yaml(self, y: 'yaml'):
        if not 'gitUserToken' in y:
            raise ValueError('missing gitUserToken')
        else:
            self.set_git_user_token(y['gitUserToken'])

        log_level = "info"
        log_path = ""
        if 'logLevel' in y:
            log_level = y['logLevel']
        if 'logPath' in y:
            log_path = y['logPath']

        self.__set_log_basic_config(log_level, log_path)

        if 'stewardServerURI' in y:
            self.set_steward_server_URI(y['stewardServerURI'])

        if 'modelRepos' in y:
            self.set_model_repos(y['modelRepos'])

        if 'tensorboardDataRootDir' in y:
            self.set_tensorboard_data_root_dir(y['tensorboardDataRootDir'])

    def __set_log_basic_config(self, log_level: str, log_path: str):
        """
        The basic config of logging can only be set once.
        """
        levels = {
            'debug': logging.DEBUG,
            'info': logging.INFO,
            'warn': logging.WARN,
            'error': logging.ERROR,
            'fatal': logging.FATAL
        }
        if log_level not in levels:
            raise ValueError(
                f"invalid log level: [{log_level}], it should be one of [debug, info, warn, error, fatal]")
        numeric_log_level = levels[log_level]
        print(f'basic log config: [level: {numeric_log_level}, filename: {log_path}]', flush=True)
        logging.basicConfig(
            level=numeric_log_level,
            filename=log_path
        )

    def set_git_user_token(self, git_user_token: str):
        self.git_user_token = git_user_token

    def set_steward_server_URI(self, uri):
        self.steward_server_URI = uri

    def set_model_repos(self, repos):
        for model_repo in repos:
            if 'gitHttpURL' not in model_repo:
                raise KeyError('invalid key of modelRepos')
            self.model_repos.append(ModelRepo(model_repo['gitHttpURL']))

    def set_tensorboard_data_root_dir(self, root_dir):
        self.tensorboard_data_root_dir = root_dir

cfg = Config()
