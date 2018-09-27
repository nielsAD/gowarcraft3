# Author:  Niels A.D.
# Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
# License: Mozilla Public License, v2.0

STORMLIB_FLAGS=-DBUILD_SHARED_LIBS=FALSE
BNCSUTIL_FLAGS=

DIR_STORMLIB=StormLib
DIR_BNCSUITL=bncsutil

CMAKE=cmake
CMAKE_FLAG=

ifeq ($(OS),Windows_NT)
	CMAKE_FLAGS+=-G'MSYS Makefiles'
endif

LIBS=$(DIR_STORMLIB)/build/libstorm.a $(DIR_BNCSUITL)/build/libbncsutil_static.a

.PHONY: all clean
all: $(LIBS)

%/build:
	mkdir -p $@

$(DIR_STORMLIB)/build/libstorm.a: $(DIR_STORMLIB)/build
	cd $(DIR_STORMLIB)/build; $(CMAKE) $(CMAKE_FLAGS) .. $(STORMLIB_FLAGS)
	$(MAKE) -C $(DIR_STORMLIB)/build

$(DIR_BNCSUITL)/build/libbncsutil_static.a: $(DIR_BNCSUITL)/build
	cd $(DIR_BNCSUITL)/build; $(CMAKE) $(CMAKE_FLAGS) .. $(BNCSUTIL_FLAGS)
	$(MAKE) -C $(DIR_BNCSUITL)/build

clean:
	-rm -r $(DIR_STORMLIB)/build
	-rm -r $(DIR_BNCSUITL)/build
