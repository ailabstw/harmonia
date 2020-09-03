# Create admin

gitea admin create-user --admin \
    --username gitea \
    --password password \
    --email admin@admin.com

# Create users

gitea admin create-user \
    --username aggregator \
    --password 1qaz_WSX \
    --email aggregator@aggregator.com \
    --must-change-password=false

gitea admin create-user \
    --username edge1 \
    --password 1qaz_WSX \
    --email edge1@edge1.com \
    --must-change-password=false

gitea admin create-user \
    --username edge2 \
    --password 1qaz_WSX \
    --email edge2@edge2.com \
    --must-change-password=false

gitea admin create-user \
    --username logserver \
    --password 1qaz_WSX \
    --email logserver@logserver.com \
    --must-change-password=false

# Create repositories

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"name": "train-plan", "auto_init": true}' \
    http://gitea:password@127.0.0.1:3000/api/v1/user/repos

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"name": "global-model", "auto_init": true}' \
    http://gitea:password@127.0.0.1:3000/api/v1/user/repos

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"name": "local-model1", "auto_init": true}' \
    http://gitea:password@127.0.0.1:3000/api/v1/user/repos

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"name": "local-model2", "auto_init": true}' \
    http://gitea:password@127.0.0.1:3000/api/v1/user/repos

# Create permissions

# train-plan
curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/train-plan/collaborators/aggregator

curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/train-plan/collaborators/edge1

curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/train-plan/collaborators/edge2

# global-model
curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "write"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/global-model/collaborators/aggregator

curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/global-model/collaborators/edge1

curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/global-model/collaborators/edge2

curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/global-model/collaborators/logserver

# local-model1
curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model1/collaborators/aggregator

curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "write"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model1/collaborators/edge1

curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model1/collaborators/logserver

# local-model2
curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model2/collaborators/aggregator

curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "write"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model2/collaborators/edge2

curl -X PUT \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"permission": "read"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model2/collaborators/logserver

# Create webhooks

# train-plan
curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://mnist-aggregator:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/train-plan/hooks

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://mnist-edge1:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/train-plan/hooks

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://mnist-edge2:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/train-plan/hooks

# global-model
curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://mnist-edge1:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/global-model/hooks

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://mnist-edge2:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/global-model/hooks

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://logserver:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/global-model/hooks

# local-model1
curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://mnist-aggregator:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model1/hooks

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://logserver:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model1/hooks

# local-model2
curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://mnist-aggregator:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model2/hooks

curl -X POST \
    -H "accept: application/json" \
    -H "Content-Type: application/json" \
    -d '{"active": true, "config": {"content_type": "json", "url": "http://logserver:9080"}, "events": ["push"], "type": "gitea"}' \
    http://gitea:password@127.0.0.1:3000/api/v1/repos/gitea/local-model2/hooks
