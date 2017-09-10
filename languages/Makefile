langs = lua python2 python3 forth bash cpp javascript php typescript
.PHONY: $(langs)
all: $(langs)

$(langs):
	docker build -t openrepl/$@ $@

list:
	@echo $(langs)
