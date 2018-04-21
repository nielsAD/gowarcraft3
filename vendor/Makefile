# Author:  Niels A.D.
# Project: gowarcraft3 (https://github.com/nielsAD/gowarcraft3)
# License: Mozilla Public License, v2.0

STORMLIB_FLAGS=-DBUILD_SHARED_LIBS=FALSE

DIR_STORMLIB=StormLib

CMAKE=cmake

LIBS=$(DIR_STORMLIB)/build/libstorm.a
all: $(LIBS)

%/build:
	mkdir -p $@

$(DIR_STORMLIB)/build/libstorm.a: $(DIR_STORMLIB)/build
	cd $(DIR_STORMLIB)/build; $(CMAKE) .. $(STORMLIB_FLAGS)
	$(MAKE) -C $(DIR_STORMLIB)/build

clean:
	rm -r $(DIR_STORMLIB)/build