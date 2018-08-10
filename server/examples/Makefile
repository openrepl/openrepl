all: docker

.PHONY: docker

docker: vendor
	docker build -t openrepl/examples .

vendor: glide.yaml
	glide up
