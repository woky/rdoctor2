TARGETS = linux.amd64 darwin.amd64
VERSION = $(shell git describe --tags --always)
OUT     = out/$(VERSION)

.PHONY: all
all: $(patsubst %, $(OUT)/%/rdoctor, $(TARGETS))

$(OUT)/version.txt:
	mkdir -p $(dir $@)
	echo $(VERSION) >$@
	rm -f out/latest
	ln -srf $(dir $@) out/latest

export GOOS   = $(word 1,$(subst ., ,$*))
export GOARCH = $(word 2,$(subst ., ,$*))
$(OUT)/%/rdoctor: $(OUT)/version.txt $(wildcard *.go)
	mkdir -p $(dir $@)
	go build -o $@ -ldflags="-X main.Version=$(VERSION)"

.PHONY: clean
clean:
	rm -rf out

-include Makefile.mine
