langs = lua python2 python3 forth bash cpp
.PHONY: $(langs)
all: $(langs)

$(langs):
	docker build -t openrepl/$@ $@

list:
	@echo $(langs)
