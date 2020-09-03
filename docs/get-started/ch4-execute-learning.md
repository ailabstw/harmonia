# Start Traininig

Push the pretrained model to `global-model` at master branch.

Push a training plan, named `plan.json`, to `train-plan` repository.
```json
{
    "name": "MNIST",
    "round": 100,
    "edge": 2,
    "EpR": 1,
    "timeout": 86400,
    "pretrainedModel": "master"
}
```

There will be 100 commits in `global-model`, `local-model1` and `local-model2` respectively. And a tag will be on the last commit of `global-model`.