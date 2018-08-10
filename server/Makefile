all: proxy runcontainer store examples

.PHONY: proxy runcontainer store examples

runcontainer:
	$(MAKE) -C runcontainer

proxy:
	$(MAKE) -C proxy

store:
	docker build -t openrepl/store store

examples:
	docker build -t openrepl/examples examples
