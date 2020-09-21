PROJ_ROOT=$(realpath ../..)

all: go_protos python_protos

go_protos:
	docker run --rm \
		-v $(PROJ_ROOT)/src/protos:/protos \
		-w /protos \
		golang:1.15.2-alpine3.12 sh -c "apk add --no-cache make && make build_go_protos"

build_go_protos:
	mkdir -p go_protos
	apk add --no-cache git protobuf make
	go get github.com/golang/protobuf/protoc-gen-go
	protoc ./service.proto -I . --go_out=plugins=grpc:./go_protos

python_protos:
	docker run --rm \
    	-v $(PROJ_ROOT)/src/protos:/protos \
		-w /protos \
    	python make build_python_protos

build_python_protos:
	mkdir -p python_protos
	python3 -m pip install --user grpcio-tools
	python3 -m grpc_tools.protoc -I . --python_out=./python_protos --grpc_python_out=./python_protos service.proto

clean:
	rm -rf go_protos python_protos