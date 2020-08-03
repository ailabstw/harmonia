# Operator Configuration

This is the minimal configurations to operators.
`type` of operator(`aggregator` or `edge`) determines the role of the pod and which repositories it should watch and operate.

Notice: field `GitHttpURL` are formated by `http[s]://<username>@<registry location>/<repository owner>/<repository name>.git`

```yml
# configs.yml

# Aggregator Config
apiVersion: v1
kind: ConfigMap
metadata:
  name: aggregator-config
data:
  aggregator-config.yml: |
    type: aggregator

    gitUserToken: <aggregator_token_generated_in_ch1>
    aggregatorModelRepo:
      gitHttpURL: http://aggregator@harmonia-gitea:3000/gitea/global-model.git
    edgeModelRepos:
      - gitHttpURL: http://aggregator@harmonia-gitea:3000/gitea/local-model1.git
      - gitHttpURL: http://aggregator@harmonia-gitea:3000/gitea/local-model2.git
    trainPlanRepo:
      gitHttpURL: http://aggregator@harmonia-gitea:3000/gitea/train-plan.git

---
# Edge1 Config

apiVersion: v1
kind: ConfigMap
metadata:
  name: edge1-config
data:
  edge-config.yml: |
    type: edge

    gitUserToken: <edge1_token_generated_in_ch1>
    aggregatorModelRepo:
      gitHttpURL: http://edge1@harmonia-gitea:3000/gitea/global-model.git
    edgeModelRepo:
      gitHttpURL: http://edge1@harmonia-gitea:3000/gitea/local-model1.git
    trainPlanRepo:
      gitHttpURL: http://edge1@harmonia-gitea:3000/gitea/train-plan.git

---
# Edge2 Config

apiVersion: v1
kind: ConfigMap
metadata:
  name: edge2-config
data:
  edge-config.yml: |
    type: edge

    gitUserToken: <edge2_token_generated_in_ch1>
    aggregatorModelRepo:
      gitHttpURL: http://edge2@harmonia-gitea:3000/gitea/global-model.git
    edgeModelRepo:
      gitHttpURL: http://edge2@harmonia-gitea:3000/gitea/local-model2.git
    trainPlanRepo:
      gitHttpURL: http://edge2@harmonia-gitea:3000/gitea/train-plan.git
```

Apply the config by
```bash
kubectl apply -f configs.yml
```