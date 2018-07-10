# Author:  Niels A.D.
# Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
# License: Mozilla Public License, v2.0

GO_FLAGS=
GOTEST_FLAGS=-cover -cpu=1,2,4 -timeout=2m

GO=go
GOFMT=gofmt
GOLINT=golint

DIR_BIN=bin
PKG:=$(shell $(GO) list ./...)
DIR:=$(subst github.com/nielsAD/gowarcraft3/,,$(PKG))

ARCH:=$(shell $(GO) env GOARCH)
ifeq ($(ARCH),amd64)
	TEST_RACE=1
endif

ifeq ($(TEST_RACE),1)
	TEST_FLAGS+= -race
endif

.PHONY: all release build test fmt lint vet list clean
all: test release

$(DIR_BIN):
	mkdir -p $@

vendor/%:
	$(MAKE) -C vendor $(subst vendor/,,$@)

release: build $(DIR_BIN)
	cd $(DIR_BIN); $(GO) list ../cmd/... | xargs -L1 go build $(GO_FLAGS)

build: vendor/StormLib/build/libstorm.a vendor/bncsutil/build/libbncsutil_static.a
	$(GO) build $(PKG)

test: build fmt lint vet
	$(GO) test $(TEST_FLAGS) $(PKG)

fmt:
	$(GOFMT) -l $(DIR)

lint:
	$(GOLINT) -set_exit_status $(PKG)

vet:
	$(GO) vet $(PKG)

list:
	@echo $(PKG) | tr ' ' '\n'

clean:
	$(MAKE) -C vendor clean
	rm -r $(DIR_BIN)
