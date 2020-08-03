CASES = \
	integration-test \
	fedavg \

all:
	for DIR in $(CASES); do \
		make -C $$DIR test clean; \
	done
