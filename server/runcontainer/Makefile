all: docker

.PHONY: docker

docker: vendor
	docker build -t openrepl/runcontainer .

vendor: glide.yaml
	glide up
