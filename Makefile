# use GNU standard variable names: https://www.gnu.org/prep/standards/html_node/Directory-Variables.html
DESTDIR :=
prefix := /usr/local
exec_prefix := $(prefix)
bindir := $(exec_prefix)/bin

# builds all the outputs
.PHONY: all
all: internal external

.PHONY: internal
internal: target/htvend target/htvend-buildah-build

.PHONY: external
external: target/htvend-buildah

# copy them to /usr/local/bin - normally run with sudo
.PHONY: install
install: all
	cp -t "$(DESTDIR)${bindir}" target/*

# remove any untracked files
.PHONY: clean
clean:
	git clean -xfd

# builds all the go binaries
target/htvend target: cmd/*/*.go internal/*/*.go go.mod go.sum
	env GOBIN=$(PWD)/target go install ./cmd/...

# copies utility scripts
target/htvend-buildah-build: scripts/htvend-buildah-build
	mkdir -p target
	cp -t target scripts/*

# copy in buildah binary (until patches are merged upstream)
target/htvend-buildah: ext-vendor/buildah/bin/buildah
	mkdir -p target
	cp ext-vendor/buildah/bin/buildah target/htvend-buildah

# build our buildah branch
ext-vendor/buildah/bin/buildah: ext-vendor/buildah
	$(MAKE) -C ext-vendor/buildah bin/buildah

# get buildah or update/clean branch that we need (until these are merged upstream)
ext-vendor/buildah:
	mkdir -p ext-vendor
	test -d ext-vendor/buildah || git -C ext-vendor clone -b continusecbuild --single-branch https://github.com/aeijdenberg/buildah.git
	git -C ext-vendor/buildah fetch
	git -C ext-vendor/buildah reset --hard origin/continusecbuild
	git -C ext-vendor/buildah clean -xfd

# ========================================================
# Following targets operate to each directory in examples/
# ========================================================
EXAMPLES := $(wildcard examples/*/)

.PHONY : manifests assets images

images: all $(addsuffix img.tar,$(EXAMPLES))
assets: all $(addsuffix assets,$(EXAMPLES))
manifests: all $(addsuffix blobs.yml,$(EXAMPLES))

%/blobs.yml: %/Dockerfile
	rm -f "$@" && env -C "$*" PATH=$(PWD)/target:$(PATH) \
		htvend build -- \
			htvend-buildah-build

%/assets: %/blobs.yml
	rm -rf "$@"
	env -C "$*" PATH=$(PWD)/target:$(PATH) \
		htvend verify --fetch
	env -C "$*" PATH=$(PWD)/target:$(PATH) \
		htvend export

%/img.tar: %/blobs.yml %/assets
	rm -f "$@" && env -C "$*" PATH=$(PWD)/target:$(PATH) \
		htvend offline --blobs-dir ./assets -- \
			htvend-buildah-build --tag oci-archive:img.tar

%/assets: %/blobs.yml
	rm -rf "$@"
	env -C "$*" PATH=$(PWD)/target:$(PATH) \
		htvend verify --fetch
	env -C "$*" PATH=$(PWD)/target:$(PATH) \
		htvend export


.PHONY: sha256sum
sha256sum: images
	sha256sum examples/*/img.tar

.PHONY: clean-examples
clean-examples:
	git -C examples clean -xfd

.PHONY: check-license
check-license:
	git ls-files | grep .go$$ | xargs go-license --config ./config/license.yml --verify

.PHONY: update-license
update-license:
	git ls-files | grep .go$$ | xargs go-license --config ./config/license.yml

.PHONY: test

test:
	go test ./...
