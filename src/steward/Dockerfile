FROM alpine

RUN apk add --no-cache git git-lfs file && \ 
    # pip3 install --no-cache-dir dvc 
    git lfs install

WORKDIR /app
ADD steward /app/steward

ENTRYPOINT ./steward
