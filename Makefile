# use GNU standard variable names: https://www.gnu.org/prep/standards/html_node/Directory-Variables.html
DESTDIR :=
prefix := /usr/local
exec_prefix := $(prefix)
bindir := $(exec_prefix)/bin

# go install to install to here
BUILDBINDIR ?= $(PWD)/target

# builds all the outputs
.PHONY: all
all: $(BUILDBINDIR)/htvend $(BUILDBINDIR)/with-temp-dir $(BUILDBINDIR)/pem2jks $(BUILDBINDIR)/build-img-with-proxy

# copy them to /usr/local/bin - normally run with sudo
.PHONY: install
install: all
	cp -t "$(DESTDIR)$(bindir)" $(BUILDBINDIR)/htvend $(BUILDBINDIR)/with-temp-dir $(BUILDBINDIR)/pem2jks $(BUILDBINDIR)/build-img-with-proxy

# remove any untracked files
.PHONY: clean
clean:
	git clean -xfd

# builds all the go binaries
$(BUILDBINDIR)/htvend $(BUILDBINDIR)/with-temp-dir $(BUILDBINDIR)/pem2jks $(BUILDBINDIR): cmd/*/*.go internal/*/*.go go.*
	GOBIN=$(BUILDBINDIR) go install -trimpath -ldflags=-buildid= ./cmd/...

# copy other scripts
$(BUILDBINDIR)/build-img-with-proxy: $(BUILDBINDIR) scripts/build-img-with-proxy
	cp scripts/build-img-with-proxy $(BUILDBINDIR)/build-img-with-proxy

.PHONY: check-license
check-license:
	git ls-files | grep .go$$ | xargs go-license --config ./config/license.yml --verify

.PHONY: update-license
update-license:
	git ls-files | grep .go$$ | xargs go-license --config ./config/license.yml

.PHONY: test
test:
	go test ./...

# builds htvend then use that to produce bootstrap assets.json for self
assets.json: all
	# here we set a temp GOMODCACHE to ensure go pulls through all dependent modules
	# we set a different BUILDBINDIR so that we won't overwrite ourselves
	$(BUILDBINDIR)/htvend build --clean -- \
		$(BUILDBINDIR)/with-temp-dir -e GOMODCACHE -e BUILDBINDIR -v -- \
			$(MAKE) -B targets-for-offline || rm assets.json

# fetch all the assets referred to by assets.json
.PHONY: fetch
fetch: assets.json $(BUILDBINDIR)/htvend 
	$(BUILDBINDIR)/htvend verify --fetch

# export all the assets referred to by assets.json to local blobs/ dir
blobs: assets.json $(BUILDBINDIR)/htvend
	rm -rf blobs/
	$(BUILDBINDIR)/htvend export

# rebuild htvend using itself and downloaded assets
.PHONY: offline
offline: $(BUILDBINDIR)/htvend $(BUILDBINDIR)/with-temp-dir assets.json blobs
	# there's no need to set GOMODCACHE, other than to demonstrate that these will be downloaded again
	$(BUILDBINDIR)/htvend offline -- \
		$(BUILDBINDIR)/with-temp-dir -e GOMODCACHE -e BUILDBINDIR -- \
			$(MAKE) BUILDBINDIR=$(BUILDBINDIR) -B targets-for-offline

.PHONY: targets-for-offline
targets-for-offline: all test

# ========================================================
# Following targets operate to each directory in examples/
# ========================================================
EXAMPLES := $(wildcard examples/*/)

.PHONY : manifests assets images

img-tarballs: $(addsuffix img.tar,$(EXAMPLES))
img-blobs: $(addsuffix blobs,$(EXAMPLES))
img-manifests: $(addsuffix assets.json,$(EXAMPLES))

%/assets.json: %/Dockerfile %/Makefile $(BUILDBINDIR)/*
	rm -f "$@"
	$(BUILDBINDIR)/htvend -C "$*" build -- \
		$(MAKE) PATH=$(BUILDBINDIR):$(PATH) -B

%/blobs: %/assets.json $(BUILDBINDIR)/htvend
	rm -rf "$@"
	$(BUILDBINDIR)/htvend -C "$*" verify --fetch
	$(BUILDBINDIR)/htvend -C "$*" export

%/img.tar: %/assets.json %/blobs $(BUILDBINDIR)/*
	rm -f "$@"
	$(BUILDBINDIR)/htvend -C "$*" offline \
		$(MAKE) PATH=$(BUILDBINDIR):$(PATH) BUILDAH_ARGS="--tag oci-archive:img.tar" -B
