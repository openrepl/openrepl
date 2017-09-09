all: languages server

.PHONY: languages server

languages:
	$(MAKE) -C languages

server:
	$(MAKE) -C server
