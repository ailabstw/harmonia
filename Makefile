IMAGE_FOLDERS = \
	src/steward \

.PHONY: all test ${IMAGE_FOLDERS}

all: images

images:
	make $(IMAGE_FOLDERS)
$(IMAGE_FOLDERS):
	make -C $@ image

clean:
	for DIR in $(IMAGE_FOLDERS) $(SDK_FOLDERS); do \
		make -C $$DIR clean; \
	done

unit-test:
	make -C src/steward test

test:
	make -C $@
