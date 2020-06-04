PROJ_ROOT=$(realpath ../..)

go-protos:
	mkdir -p protos
	docker run --rm \
		-v $(PROJ_ROOT)/src:/src \
		-w /src \
		grpc/go protoc protos/service.proto --go_out=plugins=grpc:./steward/protos

	# TODO: Update protoc-gen-go
	# docker run --rm \
	# 	-v $(PROJ_ROOT)/src:/src \
	# 	-w /src \
	# 	golang bash ./steward/protos_build.sh

image: go-protos
	docker build --tag harmonia/operator .

clean:
	rm -rf protos/ steward
	docker rmi harmonia/operator || true
