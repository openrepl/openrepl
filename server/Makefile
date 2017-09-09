all: interactive 60s proxy save

.PHONY: interactive 60s proxy save

interactive:
	docker build -t openrepl/interactive interactive

60s:
	docker build -t openrepl/60s 60s

save:
	docker build -t openrepl/save save

proxy:
	$(MAKE) -C proxy
