# Setup

Before running Harmonia FL projects, a Gitea service should be started up as a centralized git registry. Following shows how to start up Gitea in k8s. See [Gitea document](https://docs.gitea.io/en-us/install-with-docker/) for detail.

1. Gitea Environment Variable
```yml
# gitea_config.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gitea-config
data:
  SSH_DOMAIN: harmonia-gitea
  ROOT_URL: http://harmonia-gitea:3000
```

2. K8S Deployment
```yml
# gitea_deployment.yml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gitea-deployment
  labels:
    app: gitea
spec:
  replicas: 1
  selector:
    matchLabels:
      app: gitea
  template:
    metadata:
      labels:
        app: gitea
    spec:
      containers:
        - name: gitea
          image: gitea/gitea
          ports:
            - containerPort: 3000
              name: gitea
            - containerPort: 22
              name: git-ssh
          envFrom:
            - configMapRef:
                name: gitea-config
---
kind: Service
apiVersion: v1
metadata:
  name: harmonia-gitea
spec:
  selector:
    app: gitea
  ports:
  - name: gitea-http
    port: 3000
    targetPort: 3000
  - name: gitea-ssh
    port: 2222
    targetPort: 22
  type: NodePort
```

3. k8s apply
`$ kubectl apply -f gitea_config.yml`
`$ kubectl apply -f gitea_deployment.yml`

4. Install Gitea By UI
    1. Set port-forwarding  
    `$ kubectl port-forward --address 0.0.0.0 service/harmonia-gitea 3000`
    2. Install Gitea  
    `127.0.0.1:3000` can be used to access Gitea.  
    Set Gitea admin account in install page by clicking any button or directly `http://127.0.0.1:3000/install`  
    ![Install](../../assets/0-1.png)