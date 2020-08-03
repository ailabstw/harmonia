PROJ_ROOT=$(realpath ../..)

test:
	make -C $(PROJ_ROOT)/src/protos python_protos
	cp -pv $(PROJ_ROOT)/src/protos/python_protos/* edge/
	cp -pv $(PROJ_ROOT)/src/protos/python_protos/* aggregator/

	./test.sh

clean:
	docker-compose down --volumes --rmi local

	docker rm -f test_gitea
	docker network rm $(shell basename $(CURDIR))

	[ -f logs.txt ] && rm logs.txt
	[ -d registry_data ] && rm -rf registry_data

	rm -rf edge/service_pb2_grpc.py edge/service_pb2.py
	rm -rf aggregator/service_pb2_grpc.py aggregator/service_pb2.py