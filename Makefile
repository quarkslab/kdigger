CMDPACKAGE=github.com/quarkslab/kdigger/commands
VERSION=$$(git describe --tags 2>/dev/null || echo dev)
GITCOMMIT=$$(git rev-parse HEAD)
BUILDERARCH=$$(uname -m)

OUTPUTNAME=kdigger

# -w disable DWARF generation
# -s disable symbol table
# just to save some space in the binary
LDFLAGS="-s -w                               \
	-X $(CMDPACKAGE).VERSION=$(VERSION)      \
	-X $(CMDPACKAGE).GITCOMMIT=$(GITCOMMIT)  \
	-X $(CMDPACKAGE).BUILDERARCH=$(BUILDERARCH)"

# if CGO_ENABLED=1, the binary will be dynamically linked, and surprisingly,
# bigger! It seems that it is because of the net package that Go is dynamically
# linking the libraries.
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags $(LDFLAGS) -o $(OUTPUTNAME)

.PHONY: lint
lint:
	golangci-lint run

# Releasing stuff
RELEASE_FOLDER=release
RELEASE_LINUX_AMD64=$(OUTPUTNAME)-linux-amd64
RELEASE_LINUX_ARM64=$(OUTPUTNAME)-linux-arm64
RELEASE_DARWIN=$(OUTPUTNAME)-darwin-amd64

.PHONY: build-all
build-all: lint
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o $(RELEASE_LINUX_AMD64)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -o $(RELEASE_LINUX_ARM64)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -o $(RELEASE_DARWIN)

.PHONY: release
release: build-all
	mkdir -p $(RELEASE_FOLDER)
	# linux amd64
	mv $(RELEASE_LINUX_AMD64) $(RELEASE_FOLDER) && \
	cd $(RELEASE_FOLDER) && \
	sha256sum $(RELEASE_LINUX_AMD64) > $(RELEASE_LINUX_AMD64).sha256 && \
	tar cvf - $(RELEASE_LINUX_AMD64) | gzip -9 - > $(RELEASE_LINUX_AMD64).tar.gz
	# linux arm64
	mv $(RELEASE_LINUX_ARM64) $(RELEASE_FOLDER) && \
	cd $(RELEASE_FOLDER) && \
	sha256sum $(RELEASE_LINUX_ARM64) > $(RELEASE_LINUX_ARM64).sha256 && \
	tar cvf - $(RELEASE_LINUX_ARM64) | gzip -9 - > $(RELEASE_LINUX_ARM64).tar.gz
	# darwin adm64
	mv $(RELEASE_DARWIN) $(RELEASE_FOLDER) && \
	cd $(RELEASE_FOLDER) && \
	sha256sum $(RELEASE_DARWIN) > $(RELEASE_DARWIN).sha256 && \
	tar cvf - $(RELEASE_DARWIN) | gzip -9 - > $(RELEASE_DARWIN).tar.gz

DEV_IMAGE_TAG=mtardy/kdigger-dev
.PHONY: run
run: build
	echo "FROM mtardy/koolbox\nCOPY kdigger /usr/local/bin/kdigger" | docker build -t $(DEV_IMAGE_TAG) -f- .
	kind load docker-image $(DEV_IMAGE_TAG)
	kubectl run kdigger-dev-tmp --rm -i --tty --image $(DEV_IMAGE_TAG) --image-pull-policy Never -- bash

.PHONY: install-linter
install-linter:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.46.2

.PHONY: clean-docker
clean-docker:
	docker image rm $(DEV_IMAGE_TAG)

.PHONY: clean
clean:
	rm kdigger

.PHONY: clean-all
clean-all: clean clean-docker
