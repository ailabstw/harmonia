PROJ_ROOT=$(realpath ../..)

test:
	make -C $(PROJ_ROOT)/src/protos python_protos
	cp -pv $(PROJ_ROOT)/src/protos/python_protos/* edge/
	cp -pv $(PROJ_ROOT)/src/protos/python_protos/* aggregator/

	./test.sh

clean:
	docker-compose down --volumes --rmi local || true

	docker rm -f test_gitea || true
	docker network rm $(shell basename $(CURDIR)) || true

	rm -rf \
		logs.txt \
		registry_data \
		edge/service_pb2_grpc.py edge/service_pb2.py \
		timeout-edge/service_pb2_grpc.py timeout-edge/service_pb2.py \
		aggregator/service_pb2_grpc.py aggregator/service_pb2.py