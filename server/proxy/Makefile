all: docker

.PHONY: web

web:
	rm -rf web
	cp -r ../../ui/html web

docker: web
	docker build -t openrepl/proxy .
