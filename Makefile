# Author:  Niels A.D.
# Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
# License: Mozilla Public License, v2.0

THIRD_PARTY=third_party/StormLib/build/libstorm.a third_party/bncsutil/build/libbncsutil.a

GO_FLAGS=
GOTEST_FLAGS=-cover -cpu=1,2,4 -timeout=2m

GO=go
GOFMT=gofmt
STATICCHECK=$(shell $(GO) env GOPATH)/bin/staticcheck

DIR_BIN=bin
DIR_PRE=github.com/nielsAD/gowarcraft3

PKG:=$(shell $(GO) list ./...)
DIR:=$(subst $(DIR_PRE),.,$(PKG))
CMD:=$(subst $(DIR_PRE)/cmd/,,$(shell $(GO) list ./cmd/...))

ifeq ($(TEST_RACE),1)
	GOTEST_FLAGS+= -race
endif

.PHONY: all release check test fmt lint vet list clean install-tools $(CMD)

all: test release
release: $(CMD)

$(DIR_BIN):
	mkdir -p $@

$(PKG): $(THIRD_PARTY)
	$(GO) build $@

$(CMD): $(THIRD_PARTY) $(DIR_BIN)
	cd $(DIR_BIN); $(GO) build $(GO_FLAGS) $(DIR_PRE)/cmd/$@

third_party/%:
	$(MAKE) -C third_party $(subst third_party/,,$@)

check: $(THIRD_PARTY)
	$(GO) build $(PKG)

test: check fmt lint vet
	$(GO) test $(GOTEST_FLAGS) $(PKG)

fmt:
	@GOFMT_OUT=$$($(GOFMT) -d $(filter-out .,$(DIR)) $(wildcard *.go) 2>&1); \
	if [ -n "$$GOFMT_OUT" ]; then \
		echo "$$GOFMT_OUT"; \
		exit 1; \
	fi

lint:
	$(STATICCHECK) $(PKG)

vet:
	$(GO) vet $(PKG)

list:
	@echo $(PKG) | tr ' ' '\n'

clean:
	-rm -r $(DIR_BIN)
	$(GO) clean $(PKG)
	$(MAKE) -C third_party clean

install-tools:
	$(GO) mod download
	grep -o '"[^"]\+"' tools.go | xargs -n1 $(GO) install
