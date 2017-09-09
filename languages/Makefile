langs = lua python
.PHONY: $(langs)
all: $(langs)

$(langs):
	docker build -t openrepl/$@ $@

list:
	@echo $(langs)
