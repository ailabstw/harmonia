import time
from datetime import datetime
import sys
import re
import json


target_logs = [
    ('aggregator-operator',
     'Repository [aggregatorModelRepo]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git}'),
    ('aggregator-operator',
     'Repository [edgeModelRepos[0]]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git}'),
    ('aggregator-operator',
     'Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git] [/repos/harmonia_admin/global-model]...'),
    ('aggregator-operator',
     'Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git] [/repos/harmonia_admin/local-model1]...'),
    ('aggregator-operator',
     'Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/train-plan.git] [/repos/harmonia_admin/train-plan]...'),
    ('aggregator-operator', 'Init Finished'),
    ('aggregator-operator', 'received request'),
    ('aggregator-operator',
     'Pulling Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/train-plan.git]...'),
    ('aggregator-operator', 'aggregateServer.idleState'),
    ('aggregator-operator', 'aggregateServer.roundStartAction'),
    ('aggregator-operator', 'aggregateServer.localTrainState'),
    ('aggregator-operator', 'received request'),
    ('aggregator-operator', 'aggregateServer.localTrainState'),
    ('aggregator-operator', 'aggregateServer.localTrainFinishAction'),
    ('aggregator-operator',
     'Pulling Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git]...'),
    ('aggregator-operator', 'Pull Succeed'),
    ('aggregator-operator', 'aggregateServer.aggregateState'),
    ('aggregator-operator', '--- On Aggregate Finish ---'),
    ('aggregator-operator', 'aggregateServer.aggregateState'),
    ('aggregator-operator', 'aggregateServer.aggregateFinishAction'),
    ('aggregator-operator',
     'Pushing Data to [http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git]...'),
    ('aggregator-operator', 'Push Succeed'),
    ('aggregator-operator', 'aggregateServer.idleState'),
    ('edge-operator',
     'Repository [aggregatorModelRepo]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git}'),
    ('edge-operator',
     'Repository [edgeModelRepo]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git}'),
    ('edge-operator',
     'Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git] [/repos/harmonia_admin/global-model]...'),
    ('edge-operator',
     'Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git] [/repos/harmonia_admin/local-model1]...'),
    ('edge-operator',
     'Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/train-plan.git] [/repos/harmonia_admin/train-plan]...'),
    ('edge-operator', 'Init Finished'),
    ('edge-operator', 'received request'),
    ('edge-operator',
     'Pulling Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/train-plan.git]...'),
    ('edge-operator', 'edge.idleState'),
    ('edge-operator', 'edge.trainStartAction'),
    ('edge-operator', 'edge.localTrainState'),
    ('edge-operator', 'edge.trainFinishAction'),
    ('edge-operator',
     'Pushing Data to [http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git]...'),
    ('edge-operator', 'Push Succeed'),
    ('edge-operator', 'edge.aggregateState'),
    ('edge-operator', 'edge.aggregatedModelReceivedAction'),
    ('edge-operator', 'edge.localTrainState'),
    ('edge-operator', 'edge.trainFinishAction'),
    ('edge-operator', 'edge.idleState'),
    ('edge-operator', 'received request'),
    ('edge-operator', 'edge.idleState'),
    ('edge-operator', 'edge.aggregatedModelReceivedAction'),
    ('edge-operator', 'edge.idleState'),
]


n_test_cases = len(target_logs)


class LogContent:
    def __init__(self, line):
        self.container_name = None
        self.datetime = None
        self.raw_msg = None

        match = re.search(
            r'^(?P<name>[\w\-]+)[\s\|]+(?P<msg>.*)', line)
        if match:
            self.container_name = match.group('name')
            json_string = match.group('msg')
            json_obj = json.loads(json_string)
            self.datetime = datetime.strptime(json_obj['timestamp'], '%Y-%m-%dT%H:%M:%S.%fZ')
            self.raw_msg = json_string


def parse_log_file(filename):
    log_contents = []
    with open(filename, 'r') as logfile:
        for line in logfile:
            if not target_logs:
                break
            parse_line(line, log_contents)

    return check_result(log_contents)

def parse_line(line, log_contents):
    for i, (container_name, target)in enumerate(target_logs):
        if container_name in line and target in line:
            log_contents.append(LogContent(line))
            target_logs.pop(i)
            break

def check_result(log_contents):
    if target_logs or not log_contents:
        print('the number of log contents is less then test cases')
        for lg in log_contents:
            print(lg.datetime,
                  lg.container_name, lg.raw_msg)

        print('It remains...')
        for tl in target_logs:
            print(tl)
        return False
    log_contents.sort(key=lambda x: (x.container_name, x.datetime))

    prev_name = log_contents[0].container_name
    prev_dt = log_contents[0].datetime
    for lc in log_contents[1:]:
        if lc.container_name == prev_name:
            if lc.datetime < prev_dt:
                print('log contents...')
                print()
                for log_content, _ in log_contents:
                    print(log_content.datetime,
                        log_content.container_name, log_content.raw_msg)
                return False
        else:
            prev_name = lc.container_name
        prev_dt = lc.datetime
    return True


if __name__ == "__main__":
    filename = sys.argv[1]
    if parse_log_file(filename):
        print('Integration Test Pass')
    else:
        sys.exit('Integration Test Fail!')
