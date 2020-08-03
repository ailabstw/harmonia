all: image

image: protos
	docker build --tag harmonia/fedavg .

protos:
	make -C ../protos python_protos
	cp -vr ../protos/python_protos/* .

clean:
	docker rmi harmonia/fedavg || true
	rm -rf service_pb2.py service_pb2_grpc.py
