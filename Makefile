IMAGE_FOLDERS = \
	src/steward \

.PHONY: all test ${IMAGE_FOLDERS}

all:
	make images

images:
	make $(IMAGE_FOLDERS)
$(IMAGE_FOLDERS):
	make -C $@ image

clean:
	for DIR in $(IMAGE_FOLDERS) $(SDK_FOLDERS); do \
		make -C $$DIR clean; \
	done

test:
	make -C $@
