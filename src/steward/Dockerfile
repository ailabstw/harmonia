FROM golang:buster

# build golang app
RUN apt-get update -y && \
    apt-get install -y git-lfs && \
    apt-get install -y file && \
    rm -rf /var/lib/apt/lists/*
RUN git lfs install
WORKDIR /app
ADD . /app

RUN cd /app && go build
ENTRYPOINT ./steward
