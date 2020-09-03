from logserver import gitoperation
import os
import unittest
from unittest.mock import patch, Mock


@patch('logserver.gitoperation.base_dir', 'repos')
class Testgitoperation(unittest.TestCase):

    @patch('logserver.gitoperation.create_cred_helper_script')
    @patch('subprocess.run')
    def test_setup_git(self, exec_func, cred_func):
        gitoperation.setup_git("repo_path", Mock())
        cred_func.assert_called_once()
        self.assertEqual(exec_func.call_count, 2)

    def test_convert_git_http_url_to_full_name(self):
        result = gitoperation.convert_git_http_url_to_full_name(
            "http://hmadmin@gitea:3000/hmadmin/local-model1.git")
        self.assertEqual(result, "hmadmin/local-model1")

    def test_get_repo_path(self):
        result = gitoperation.get_repo_path(
            "http://hmadmin@gitea:3000/hmadmin/local-model1.git")
        self.assertEqual(result, 'repos/hmadmin/local-model1')

    @patch('builtins.open')
    @patch('os.chmod')
    def test_create_cred_helper_script(self, chomd_func, open_func):
        gitoperation.create_cred_helper_script("abc")
        open_func.assert_called_once()
        chomd_func.assert_called_once()

    @patch('os.path.isdir')
    @patch('logserver.gitoperation.get_repo_path')
    @patch('logserver.gitoperation.exec_git_password_command')
    def test_clone_repo_without_existed(self, exec_func, get_func, isdir_func):
        isdir_func.return_value = False
        gitoperation.clone_repo("repo_path", "git_http_url")
        get_func.assert_called_once()
        exec_func.assert_called_once()

    @patch('os.path.isdir')
    @patch('logserver.gitoperation.get_repo_path')
    @patch('logserver.gitoperation.exec_git_password_command')
    def test_clone_repo_with_existed(self, exec_func, get_func, isdir_func):
        isdir_func.return_value = True
        gitoperation.clone_repo("repo_path", "git_http_url")
        get_func.assert_not_called()
        exec_func.assert_not_called()

    @patch('logserver.gitoperation.exec_command')
    def test_fetch_repo(self, exec_func):
        gitoperation.fetch_repo("repo_path")
        exec_func.assert_called_once()

    @patch('logserver.gitoperation.exec_git_password_command')
    def test_checkout_file(self, exec_func):
        gitoperation.checkout_file("repo_path", "tree_ish", "filename")
        exec_func.assert_called_once()

    @patch('logserver.gitoperation.credential_helper_script', 'credential')
    @patch('logserver.gitoperation.exec_command')
    def test_exec_git_password_command(self, exec_func):
        gitoperation.exec_git_password_command([], "path")
        exec_func.assert_called_once()
        git, args, path = exec_func.call_args[0]
        env = exec_func.call_args[1]
        self.assertEqual(git, 'git')
        self.assertFalse(args)
        self.assertEqual(path, 'path')
        self.assertTrue(
            'GIT_ASKPASS' in env['env'] and env['env']['GIT_ASKPASS'] == 'credential')

    @patch('subprocess.run')
    def test_exec_command(self, exec_func):
        gitoperation.exec_command('command', [])
        exec_func.assert_called_once()


if __name__ == "__main__":
    """
    verbosity:
    0 (quiet): you just get the total numbers of tests executed and the global result
    1 (default): you get the same plus a dot for every successful test or a F for every failure
    2 (verbose): you get the help string of every test and the result
    """
    unittest.main(verbosity=2)
