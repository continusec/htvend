# use GNU standard variable names: https://www.gnu.org/prep/standards/html_node/Directory-Variables.html
DESTDIR :=
prefix := /usr/local
exec_prefix := $(prefix)
bindir := $(exec_prefix)/bin

# go install to install to here
BUILDBINDIR ?= $(CURDIR)/target
GOSOURCES := $(shell git ls-files *.go)

# all binaries that we create
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

# ========================================================
# Tool image (consumed by the Bazel rules in ../rules)
# ========================================================
# Bundles htvend + build-img-with-proxy + upstream buildah on a wolfi base.
# See ./Dockerfile.
IMAGE ?= ghcr.io/continusec/htvend
IMAGE_TAG ?= dev
PLATFORMS ?= linux/amd64,linux/arm64

# build the tool image locally for the current architecture
.PHONY: image
image:
	docker build -t "$(IMAGE):$(IMAGE_TAG)" -f Dockerfile .

# build the tool image for all platforms and push to the registry
.PHONY: image-push
image-push:
	docker buildx build --platform "$(PLATFORMS)" -t "$(IMAGE):$(IMAGE_TAG)" --push -f Dockerfile .
