# use GNU standard variable names: https://www.gnu.org/prep/standards/html_node/Directory-Variables.html
DESTDIR :=
prefix := /usr/local
exec_prefix := $(prefix)
bindir := $(exec_prefix)/bin

# builds all the outputs
.PHONY: all
all: target/htvend target/with-temp-dir

# copy them to /usr/local/bin - normally run with sudo
.PHONY: install
install: all
	cp -t "$(DESTDIR)${bindir}" target/htvend target/with-temp-dir

# remove any untracked files
.PHONY: clean
clean:
	git clean -xfd

# builds all the go binaries
target/htvend target/with-temp-dir target: cmd/*/*.go internal/*/*.go go.mod go.sum
	env GOBIN=$(PWD)/target go install -trimpath -ldflags=-buildid= ./cmd/...

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
assets.json: target/htvend target/with-temp-dir go.mod go.sum
	# here we set a temp GOMODCACHE to ensure go pulls through all dependent modules
	./target/htvend build --clean -- \
		./target/with-temp-dir -e GOMODCACHE -- \
			$(MAKE) -B targets-for-offline || rm assets.json

# fetch all the assets referred to by assets.json
.PHONY: fetch
fetch: assets.json target/htvend 
	./target/htvend verify --fetch

# export all the assets referred to by assets.json to local blobs/ dir
blobs: assets.json target/htvend
	rm -rf blobs/
	./target/htvend export

# rebuild htvend using itself and downloaded assets
.PHONY: offline
offline: target/htvend target/with-temp-dir assets.json blobs
	# there's no need to set GOMODCACHE, other than to demonstrate that these will be downloaded again
	./target/htvend offline -- \
		./target/with-temp-dir -e GOMODCACHE -- \
			$(MAKE) -B targets-for-offline

.PHONY: targets-for-offline
targets-for-offline: target/htvend target/with-temp-dir test
