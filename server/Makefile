all: proxy runcontainer store

.PHONY: proxy runcontainer store

runcontainer:
	$(MAKE) -C runcontainer

proxy:
	$(MAKE) -C proxy

store:
	docker build -t openrepl/store store
