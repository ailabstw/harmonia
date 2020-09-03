import re
import logging
import os
import subprocess
import stat


base_dir = '/repos'
credential_helper_script = "/tmp/credentialHelper.sh"
logger = logging.getLogger(__name__)


class ModelRepo():
    def __init__(self, git_http_url):
        self.git_http_url = git_http_url


class GitUser:
    def __init__(self, name: str, email: str, token: str):
        self.name = name
        self.email = email
        self.token = token


def setup_git(repo_path: str, git_user: GitUser):
    logger.info('setup git user')
    create_cred_helper_script(git_user.token)
    subprocess.run(['git', 'config', 'user.email',
                    git_user.email], cwd=repo_path, check=True)
    subprocess.run(['git', 'config', 'user.name', git_user.name],
                   cwd=repo_path, check=True)


def create_cred_helper_script(user_token: str):
    logger.info('create credential helper script')
    txt = "printf '%s\\n' " + user_token
    f = open(credential_helper_script, 'w')
    f.write(txt)
    f.close()
    os.chmod(credential_helper_script, stat.S_IRUSR |
             stat.S_IWUSR | stat.S_IXUSR)


def convert_git_http_url_to_full_name(git_http_url: str) -> str:
    pattern = r"(?P<method>https?):\/\/(?:[\w_-]+@)(?P<provider>.*?(?P<port>\:\d+)?)(?:\/|:)(?P<handle>(?P<owner>.+?)\/(?P<repo>.+?))(?:\.git|\/)?$"
    result = re.match(pattern, git_http_url)
    return '{}/{}'.format(result['owner'], result['repo'])


def get_repo_path(git_http_url: str) -> str:
    return os.path.join(base_dir, convert_git_http_url_to_full_name(git_http_url))


def clone_repo(repo_path: str, git_http_url: str):
    logger.info(f'cloning repo {git_http_url}...')
    if os.path.isdir(os.path.join(repo_path, '.git')):
        logger.debug(f'repo {git_http_url} already existed at {repo_path}')
        return

    repo_path = get_repo_path(git_http_url)
    exec_git_password_command(['clone', git_http_url, repo_path])


def fetch_repo(repo_path: str):
    logger.debug(f'fetch repo {repo_path}')
    exec_command('git', ['fetch'], path=repo_path)


def checkout_file(repo_path: str, tree_ish: str, filename: str):
    logger.debug(f'checkout file {filename} from repo: {repo_path} with ref: {tree_ish}')
    exec_git_password_command(
        ['checkout', tree_ish, '--', filename], path=repo_path)


def exec_git_password_command(args, path=None):
    env = os.environ.copy()
    env['GIT_ASKPASS'] = credential_helper_script
    exec_command('git', args, path, env=env)


def exec_command(name: str, args, path=None, env=os.environ.copy()):
    logging.debug(f'exec commands: {[name] + args}')
    try:
        p = subprocess.run([name] + args, cwd=path, env=env,
                           stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    except subprocess.SubprocessError as err:
        logging.fatal(f'exec commands error: {err}')
    if p.stderr:
        logging.fatal(f'command execution stderr: {p.stderr}')
    if p.stdout:
        logging.debug(f'execution stdout: {p.stdout}')
