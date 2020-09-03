from logserver import config
import json
import logging
import os
import tempfile
import unittest
import yaml
from unittest.mock import patch, Mock


@patch.dict(config.Config._instances, {}, clear=True)
class TestConfig(unittest.TestCase):

    def setUp(self):
        config.Config._instances = {}
        self.cfg = config.Config(None)

    def test_empty_init(self):
        self.assertEqual(self.cfg.steward_server_URI, '0.0.0.0:9080')
        self.assertEqual(self.cfg.tensorboard_data_root_dir,
                         '/tensorboard_data')
        self.assertFalse(self.cfg.model_repos)

    @patch.object(config.Config, '_Config__parse_yaml')
    @patch.object(config.Config, '_Config__read_yaml')
    @patch('os.path.exists')
    def test_init(self, exists_func, read_func, parse_func):
        exists_func.return_value = True
        read_func.return_value = {}
        with self.assertRaises(ValueError):
            config.Config().__init__(Mock())
        read_func.return_value = True
        config.Config().__init__(Mock())
        parse_func.assert_called_once()

    @patch.object(config.Config, '_Config__parse_yaml')
    def test_read_yaml(self, parse_func):
        f = tempfile.NamedTemporaryFile(mode='w+')
        dct = {
            'gitUserToken': 'abc',
            'logLevel': 'warn',
            'logPath': '/a/b/c',
            'stewardServerURI': 'localhost:9999',
            'modelRepos': [
                {'gitHttpURL': 'http://a.git'},
                {'gitHttpURL': 'http://b.git'},
                {'gitHttpURL': 'http://c.git'},
            ],
            'tensorboardDataRootDir': '/data'
        }
        f.write(json.dumps(dct))
        f.seek(0)
        config.Config._instances = {}
        config.Config(f.name)
        f.close()
        parse_func.assert_called_once()
        for k in dct:
            self.assertEqual(parse_func.call_args[0][0][k], dct[k])

    @patch.object(config.Config, 'set_tensorboard_data_root_dir')
    @patch.object(config.Config, 'set_model_repos')
    @patch.object(config.Config, 'set_steward_server_URI')
    @patch.object(config.Config, '_Config__set_log_basic_config')
    @patch.object(config.Config, 'set_git_user_token')
    @patch.object(config.Config, '_Config__read_yaml')
    @patch('os.path.exists')
    def test_parse_yaml(self, exists_func, read_func, token_func,
                        logconfig_func, steward_func, repos_func, dir_func):
        exists_func.return_value = True
        yml = {
            'gitUserToken': 'abc',
            'logLevel': 'warn',
            'logPath': '/a/b/c',
            'stewardServerURI': 'localhost:9999',
            'modelRepos': [
                {'gitHttpURL': 'http://a.git'},
                {'gitHttpURL': 'http://b.git'},
                {'gitHttpURL': 'http://c.git'},
            ],
            'tensorboardDataRootDir': '/data'
        }
        read_func.return_value = yml
        config.Config().__init__(Mock())
        token_func.assert_called_once()
        logconfig_func.assert_called_once()
        steward_func.assert_called_once()
        repos_func.assert_called_once()
        dir_func.assert_called_once()
        self.assertTupleEqual(token_func.call_args[0], (yml['gitUserToken'],))
        self.assertTupleEqual(
            logconfig_func.call_args[0], (yml['logLevel'], yml['logPath']))
        self.assertTupleEqual(
            steward_func.call_args[0], (yml['stewardServerURI'],))
        self.assertTupleEqual(repos_func.call_args[0], (yml['modelRepos'],))
        self.assertTupleEqual(
            dir_func.call_args[0], (yml['tensorboardDataRootDir'],))

    def test_set_git_user_token(self):
        test_token = 'user_token'
        self.cfg.set_git_user_token(test_token)
        self.assertEqual(self.cfg.git_user_token, test_token)

    def test_set_steward_server_URI(self):
        test_uri = '127.0.0.1:8888'
        self.cfg.set_steward_server_URI(test_uri)
        self.assertEqual(self.cfg.steward_server_URI, test_uri)

    def test_set_model_repos(self):
        test_git_http_urls = set(['abc', 'cde'])
        test_repos = [
            {'gitHttpURL': url} for url in test_git_http_urls
        ]
        self.cfg.set_model_repos(test_repos)
        self.assertListEqual(sorted(
            [repo.git_http_url for repo in self.cfg.model_repos]), sorted(test_git_http_urls))

    def test_set_tensorboard_data_root_dir(self):
        test_root_dir = '/root'
        self.cfg.set_tensorboard_data_root_dir(test_root_dir)
        self.assertEqual(self.cfg.tensorboard_data_root_dir, test_root_dir)


if __name__ == "__main__":
    """
    verbosity:
    0 (quiet): you just get the total numbers of tests executed and the global result
    1 (default): you get the same plus a dot for every successful test or a F for every failure
    2 (verbose): you get the help string of every test and the result
    """
    unittest.main(verbosity=2)
