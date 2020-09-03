PROJ_ROOT=$(realpath ../..)

all: image

image: 
    # build logserver
	docker build --tag harmonia/logserver .

clean:
	docker rmi harmonia/logserver || true

unit_test:
	apt-get update
	pip install -r logserver/requirements.txt
	python -m unittest -v logserver/tests/*.py

test:
	docker run --rm \
		-v $(PROJ_ROOT)/src/logserver:/src/logserver \
		-v $(PROJ_ROOT)/src/logserver/Makefile:/src/Makefile \
		-w /src \
		python:3.8 make unit_test
