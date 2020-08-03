FROM python

RUN apt-get update -y && \
    rm -rf /var/lib/apt/lists/*
RUN pip3 install --upgrade pip && pip3 install grpcio grpcio-tools
WORKDIR /app
ADD . /app
ENTRYPOINT [ "python3", "grpc_server.py" ]
