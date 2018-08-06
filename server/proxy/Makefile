all: docker

.PHONY: web

web:
	rm -rf web
	cp -r ../../ui web

docker: web
	docker build -t openrepl/proxy .
