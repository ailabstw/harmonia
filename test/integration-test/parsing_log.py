import sys
import re

EXPECTED = {
    'aggregator-operator_1': [
        ('Repository [aggregatorModelRepo]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git}', ),
        ('Repository [edgeModelRepos[0]]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git}', ),
        ('Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git] [/repos/harmonia_admin/global-model]...', ),
        ('Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git] [/repos/harmonia_admin/local-model1]...', ),
        ('Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/train-plan.git] [/repos/harmonia_admin/train-plan]...', ),
        ('Init Finished', ),
        ('Receive webhook', 'train-plan', ),
        ('Pulling Data [/repos/harmonia_admin/train-plan]...', ),
        ('aggregateServer.idleState', ),
        ('aggregateServer.trainStartAction', ),
        ('aggregateServer.localTrainState', ),
        ('Receive webhook', 'local-model', ),
        ('aggregateServer.localTrainState', ),
        ('aggregateServer.localTrainFinishAction', ),
        ('Pulling Data [/repos/harmonia_admin/local-model1]...', ),
        ('Pull Succeed', ),
        ('aggregateServer.localTrainTimeoutAction', ),
        ('aggregateServer.aggregateState', ),
        ('--- On Aggregate Finish ---', ),
        ('aggregateServer.aggregateState', ),
        ('aggregateServer.aggregateFinishAction', ),
        ('Pushing to [/repos/harmonia_admin/global-model] args [[--all]]...', ),
        ('Push Succeed', ),
        ('aggregateServer.idleState', ),
    ],
    'push-edge-operator_1': [
        ('Repository [aggregatorModelRepo]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git}', ),
        ('Repository [edgeModelRepo]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model2.git}', ),
        ('Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git] [/repos/harmonia_admin/global-model]...', ),
        ('Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model2.git] [/repos/harmonia_admin/local-model2]...', ),
        ('Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/train-plan.git] [/repos/harmonia_admin/train-plan]...', ),
        ('Init Finished', ),
        ('Receive webhook', ),
        ('Pulling Data [/repos/harmonia_admin/train-plan]...', ),
        ('edge.idleState', 1, ),
        ('edge.trainPlanAction', ),
        ('edge.trainInitState', ),
        ('edge.trainStartAction', ),
        ('edge.localTrainState', ),
        ('edge.trainFinishAction', ),
        ('Pushing to [/repos/harmonia_admin/local-model2] args [[--all]]...', ),
        ('Push Succeed', ),
        ('edge.aggregateState', ),
        ('edge.baseModelReceivedAction', ),
        ('edge.localTrainState', ),
        ('edge.trainFinishAction', ),
        ('edge.idleState', 2, ),
        ('Receive webhook', ),
        ('edge.idleState', 3, ),
        ('edge.baseModelReceivedAction', ),
    ],
    'pull-edge-operator_1': [
        ('Repository [aggregatorModelRepo]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git}', ),
        ('Repository [edgeModelRepo]: &{http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git}', ),
        ('Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/global-model.git] [/repos/harmonia_admin/global-model]...', ),
        ('Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/local-model1.git] [/repos/harmonia_admin/local-model1]...', ),
        ('Cloning Data from [http://harmonia_admin@test_gitea:3000/harmonia_admin/train-plan.git] [/repos/harmonia_admin/train-plan]...', ),
        ('Init Finished', ),
        ('Pulling Data [/repos/harmonia_admin/train-plan]...', ),
        ('edge.idleState', 1, ),
        ('edge.trainPlanAction', ),
        ('edge.trainInitState', ),
        ('edge.trainStartAction', ),
        ('edge.localTrainState', ),
        ('edge.trainFinishAction', ),
        ('Pushing to [/repos/harmonia_admin/local-model1] args [[--all]]...', ),
        ('Push Succeed', ),
        ('edge.aggregateState', ),
        ('edge.baseModelReceivedAction', ),
        ('edge.localTrainState', ),
        ('edge.trainFinishAction', ),
        ('edge.idleState', 2, ),
        ('edge.idleState', 3, ),
        ('edge.baseModelReceivedAction', ),
    ],
    'push-edge-app_1': [
        ('Train Finish', ),
        ('Server Stop', ),
    ],
    'pull-edge-app_1': [
        ('Train Finish', ),
        ('Server Stop', ),
    ],
}

def parse_log_file(filename):
    with open(filename, 'r') as logfile:
        def docker_log_parse(line):
            try:
                match = re.search(r'^(?P<name>[\w\-]+)[\s]*\|[\s]*(?P<msg>.*)', line)
                return match.group('name'), match.group('msg')
            except Exception:
                return None, None


        parsed_logs = filter(
            lambda line: line[0],
            map(docker_log_parse, logfile),
        )
        # pprint(list(parsed_logs))

        grouped_logs = {}
        for parsed_log in parsed_logs:
            if parsed_log[0] in grouped_logs:
                grouped_logs[parsed_log[0]].append(parsed_log[1])
            else:
                grouped_logs[parsed_log[0]] = [parsed_log[1]]

        return grouped_logs

def validate(actual_logs):
    insufficient_logs = dict(map(
        lambda expected_container_logs: (
            expected_container_logs[0],
            list(filter(
                lambda expected_log: not any(map(lambda actual_log: expected_log[0] in actual_log, actual_logs[expected_container_logs[0]])),
                expected_container_logs[1],
            )),
        ),
        EXPECTED.items(),
    ))

    return insufficient_logs


def main():
    filename = sys.argv[1]
    actual_logs = parse_log_file(filename)

    assert all(map(
        lambda insufficient_container_logs: len(insufficient_container_logs[1]) == 0,
        validate(actual_logs).items()
    )), validate(actual_logs)

if __name__ == "__main__":
    main()
