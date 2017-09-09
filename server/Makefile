all: interactive 60s proxy

.PHONY: interactive 60s proxy

interactive:
	docker build -t openrepl/interactive interactive

60s:
	docker build -t openrepl/60s 60s

proxy:
	$(MAKE) -C proxy
