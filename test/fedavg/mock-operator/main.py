from concurrent import futures
from pathlib import Path
import os
import signal
import operator
import sys
import time

import torch
import grpc
import service_pb2
import service_pb2_grpc

EXPECTED = {
    'aaa': (3 * 1 + 6 * 2) / 3,
    'bbb': (6 * 1 + 3 * 2) / 3,
}

class AggregateServerServicer(service_pb2_grpc.AggregateServerOperatorServicer):
    def AggregateFinish(self, request, _):
        print(f"Aggregate finish: [{request}]")

        result = torch.load("/repos/output/weights.tar", 'cpu')
        print(f"Aggregate result: [{result}]")

        if operator.eq(result, EXPECTED):
            os.kill(os.getpid(), signal.SIGUSR1)
        else:
            print(f"Expect [{EXPECTED}], Got [{result}]")
            os.kill(os.getpid(), signal.SIGUSR2)

        sys.stdout.flush()
        return service_pb2.Empty()

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_AggregateServerOperatorServicer_to_server(
        AggregateServerServicer(), server)
    server.add_insecure_port("0.0.0.0:1111")
    server.start()

def send_message():
    print("Sending message")

    channel = grpc.insecure_channel("fedavg:2222")
    stub = service_pb2_grpc.AggregateServerAppStub(channel)

    stub.Aggregate(
        service_pb2.AggregateParams(
            localModels=[{
                "path": "input1",
                "datasetSize": 1,
            }, {
                "path": "input2",
                "datasetSize": 2,
            }],
            aggregatedModel={
                "path": "output",
            },
        )
    )

def validate(signum, _):
    sys.stdout.flush()
    if signum == signal.SIGUSR1:
        sys.exit(0)
    else:
        sys.exit(1)

if __name__ == "__main__":
    serve()

    Path("/repos/input1").mkdir(parents=True, exist_ok=True)
    Path("/repos/input2").mkdir(parents=True, exist_ok=True)
    Path("/repos/output").mkdir(parents=True, exist_ok=True)

    torch.save({
        'aaa': 3,
        'bbb': 6,
    }, "/repos/input1/weights.tar")

    torch.save({
        'aaa': 6,
        'bbb': 3,
    }, "/repos/input2/weights.tar")


    signal.signal(signal.SIGUSR1, validate)
    signal.signal(signal.SIGUSR2, validate)

    send_message()

    while True:
        time.sleep(100)
