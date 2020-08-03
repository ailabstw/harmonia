FROM pytorch/pytorch

RUN pip install -U pip && \
    pip install grpcio grpcio-tools

ADD . /app
ENTRYPOINT [ "python3", "/app/main.py" ]
