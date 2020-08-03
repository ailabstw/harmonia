PROJ_ROOT=$(realpath ../..)

test:
	make -C $(PROJ_ROOT)/src/protos python_protos
	cp -pv $(PROJ_ROOT)/src/protos/python_protos/* mock-operator

	docker-compose up --build -d

	test `docker wait $(notdir $(CURDIR))_validator_1` = 0

clean:
	docker-compose down --rmi local

	rm -rf repos
	rm -rf mock-operator/service_pb2.py mock-operator/service_pb2_grpc.py 