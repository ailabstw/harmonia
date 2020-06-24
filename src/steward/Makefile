PROJ_ROOT=$(realpath ../..)

all: image

protos:
	mkdir -p protos
	docker run --rm \
		-v $(PROJ_ROOT)/src:/src \
		-w /src \
		grpc/go protoc protos/service.proto --go_out=plugins=grpc:./steward/protos

protos-v3:
	# TODO: Update protoc-gen-go
	# docker run --rm \
	# 	-v $(PROJ_ROOT)/src:/src \
	# 	-w /src \
	# 	golang bash ./steward/protos_build.sh

steward: protos
	docker run --rm \
		-v $(PROJ_ROOT)/src/steward:/src \
		-w /src \
		golang:buster go build

image: steward 
	docker build --tag harmonia/operator .

clean:
	rm -rf protos/ steward
	docker rmi harmonia/operator || true

test: protos
	go test ./...