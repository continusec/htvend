# use GNU standard variable names: https://www.gnu.org/prep/standards/html_node/Directory-Variables.html
DESTDIR :=
prefix := /usr/local
exec_prefix := $(prefix)
bindir := $(exec_prefix)/bin

# builds all the outputs
.PHONY: all
all: target/htvend

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
	env GOBIN=$(PWD)/target go install ./cmd/...

.PHONY: check-license
check-license:
	git ls-files | grep .go$$ | xargs go-license --config ./config/license.yml --verify

.PHONY: update-license
update-license:
	git ls-files | grep .go$$ | xargs go-license --config ./config/license.yml

.PHONY: test
test:
	go test ./...

# builds htvend then use that to produce bootstrap for self
blobs.yml: target/htvend target/with-temp-dir
	./target/htvend build --clean -- \
		./target/with-temp-dir -e GOMODCACHE -- \
			$(MAKE) -B target/htvend || rm blobs.yml
