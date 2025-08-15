# use GNU standard variable names: https://www.gnu.org/prep/standards/html_node/Directory-Variables.html
DESTDIR :=
prefix := /usr/local
exec_prefix := $(prefix)
bindir := $(exec_prefix)/bin

# go install to install to here
BUILDBINDIR ?= $(CURDIR)/target
GOSOURCES := $(shell git ls-files *.go)

# all binaries that we creaate
go_bins := $(patsubst cmd/%,$(BUILDBINDIR)/%,$(wildcard cmd/*))
scripts := $(patsubst scripts/%,$(BUILDBINDIR)/%,$(wildcard scripts/*))

all_artifacts := $(go_bins) $(scripts)

# builds all the outputs
.PHONY: all
all: $(all_artifacts)

# copy them to /usr/local/bin - normally run with sudo
.PHONY: install
install: $(all_artifacts)
	cp -t "$(DESTDIR)$(bindir)" $(all_artifacts)

# remove any untracked files
.PHONY: clean
clean:
	git clean -xfd

$(go_bins) $(BUILDBINDIR): $(GOSOURCES) go.*
	GOBIN=$(BUILDBINDIR) CGO_ENABLED=0 go install -trimpath -ldflags=-buildid= ./cmd/...

# copy other scripts
$(scripts): $(go_bins) scripts/* $(BUILDBINDIR)
	cp -t $(BUILDBINDIR) scripts/*

.PHONY: check-license
check-license:
	git ls-files | grep .go$$ | xargs go-license --config ./config/license.yml --verify

.PHONY: update-license
update-license:
	git ls-files | grep .go$$ | xargs go-license --config ./config/license.yml

.PHONY: test
test:
	go test ./...

.PHONY: targets-for-offline
targets-for-offline: all test

# builds htvend then use that to produce bootstrap assets.json for self
assets.json: $(all_artifacts)
	# here we set a temp GOMODCACHE to ensure go pulls through all dependent modules
	# we set a different BUILDBINDIR so that we won't overwrite ourselves
	$(BUILDBINDIR)/htvend build --clean -t GOMODCACHE -t BUILDBINDIR -- \
		$(MAKE) -B targets-for-offline || rm assets.json

# fetch all the assets referred to by assets.json
blobs: assets.json $(all_artifacts)
	rm -rf blobs
	$(BUILDBINDIR)/htvend verify --fetch
	$(BUILDBINDIR)/htvend export

# rebuild htvend using itself and downloaded assets
.PHONY: offline
offline: $(all_artifacts) assets.json blobs
	# there's no need to set GOMODCACHE, other than to demonstrate that these will be downloaded again
	$(BUILDBINDIR)/htvend offline -t GOMODCACHE -t BUILDBINDIR -- \
		$(MAKE) BUILDBINDIR=$(BUILDBINDIR) -B targets-for-offline

# clone or update patched-buildah-src
patched-buildah-src:
	test -d patched-buildah-src || git clone https://github.com/aeijdenberg/buildah --branch continusecbuild --single-branch patched-buildah-src
	git -C patched-buildah-src fetch
	git -C patched-buildah-src checkout continusecbuild
	git -C patched-buildah-src reset --hard origin/continusecbuild

# build patched buildah binary
patched-buildah-src/bin/buildah: patched-buildah-src
	$(MAKE) -C patched-buildah-src bin/buildah

# run with sudo to install patched-buildah binary and init config files for it
.PHONY: install-patched-buildah
install-patched-buildah: patched-buildah-src/bin/buildah
	# copy our patched binary
	cp patched-buildah-src/bin/buildah "$(DESTDIR)$(bindir)/patched-buildah"

	# the following are normally installed by podman / buildah. If they don't exist, then drop default values in
	mkdir -p /etc/containers
	test -f /etc/containers/registries.conf || (echo 'unqualified-search-registries = ["docker.io"]' > /etc/containers/registries.conf)
	test -f /etc/containers/policy.json || (echo '{"default":[{"type":"insecureAcceptAnything"}]}' > /etc/containers/policy.json)

# ========================================================
# Following targets operate to each directory in examples/
# ========================================================
EXAMPLES := $(wildcard examples/*/)

.PHONY : img-tarballs img-blobs img-manifests
img-tarballs: $(addsuffix img.tar,$(EXAMPLES))
img-blobs: $(addsuffix blobs,$(EXAMPLES))
img-manifests: $(addsuffix assets.json,$(EXAMPLES))

%/assets.json: %/Dockerfile %/Makefile $(all_artifacts)
	rm -f "$@"
	$(BUILDBINDIR)/htvend -C "$*" build -- \
		$(MAKE) PATH=$(BUILDBINDIR):$(PATH) -B

%/blobs: %/assets.json $(all_artifacts)
	rm -rf "$@"
	$(BUILDBINDIR)/htvend -C "$*" verify --fetch
	$(BUILDBINDIR)/htvend -C "$*" export

%/img.tar: %/assets.json %/blobs $(all_artifacts)
	rm -f "$@"
	$(BUILDBINDIR)/htvend -C "$*" offline -- \
		$(MAKE) PATH=$(BUILDBINDIR):$(PATH) BUILDAH_OPTS="--tag oci-archive:img.tar" -B

.PHONY: sha256sums
sha256sums: img-tarballs
	sha256sum examples/*/img.tar
