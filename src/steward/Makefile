PROJ_ROOT=$(realpath ../..)

all: image

protos:
	make -C ../protos go_protos
	cp -pvr ../protos/go_protos protos

image: protos
	# build steward
	docker run --rm \
		-v $(PROJ_ROOT)/src/steward:/src \
		-w /src \
		golang:1.15.2-alpine3.12 go build
	docker build --tag harmonia/operator .

debug_image: protos
	docker build --file Dockerfile_debug --tag harmonia/operator .

clean:
	rm -rf protos/ steward
	docker rmi harmonia/operator || true

unit_test:
	apt update
	apt-get install -y git-lfs file
	git lfs install
	go test ./... -v

test: protos
	docker run --rm \
		-v $(PROJ_ROOT)/src/steward:/src \
		-w /src \
		golang:buster make unit_test
