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
    gitUserToken: aggregator_password_or_token
    aggregatorModelRepo:
      gitHttpURL: http://aggregator@harmonia-gitea:3000/hmadmin/Aggregated-Model.git
    edgeModelRepos:
      - gitHttpURL: http://aggregator@harmonia-gitea:3000/hmadmin/Edge1-Model.git
      - gitHttpURL: http://aggregator@harmonia-gitea:3000/hmadmin/Edge2-Model.git
    trainPlanRepo:
      gitHttpURL: http://aggregator@harmonia-gitea:3000/hmadmin/Train-Plan.git

---
# Edge1 Config

apiVersion: v1
kind: ConfigMap
metadata:
  name: edge1-config
data:
  edge-config.yml: |
    type: edge
    gitUserToken: edge1_password_or_token
    aggregatorModelRepo:
      gitHttpURL: http://edge1@harmonia-gitea:3000/hmadmin/Aggregated-Model.git
    edgeModelRepo:
      gitHttpURL: http://edge1@harmonia-gitea:3000/hmadmin/Edge1-Model.git
    trainPlanRepo:
      gitHttpURL: http://edge1@harmonia-gitea:3000/hmadmin/Train-Plan.git

---
# Edge2 Config

apiVersion: v1
kind: ConfigMap
metadata:
  name: edge2-config
data:
  edge-config.yml: |
    type: edge
    gitUserToken: edge2_password_or_token
    aggregatorModelRepo:
      gitHttpURL: http://edge2@harmonia-gitea:3000/hmadmin/Aggregated-Model.git
    edgeModelRepo:
      gitHttpURL: http://edge2@harmonia-gitea:3000/hmadmin/Edge2-Model.git
    trainPlanRepo:
      gitHttpURL: http://edge2@harmonia-gitea:3000/hmadmin/Train-Plan.git
```

Apply the config by
`$ kubectl apply -f configs.yml`