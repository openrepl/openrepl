all: interactive 60s proxy save runcontainer

.PHONY: interactive 60s proxy save runcontainer

interactive:
	docker build -t openrepl/interactive interactive

60s:
	docker build -t openrepl/60s 60s

save:
	docker build -t openrepl/save save

runcontainer:
	docker build -t openrepl/runcontainer runcontainer

proxy:
	$(MAKE) -C proxy
