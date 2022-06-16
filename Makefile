CMDPACKAGE=github.com/quarkslab/kdigger/commands
VERSION=$$(git describe --tags 2>/dev/null || echo dev)
GITCOMMIT=$$(git rev-parse HEAD)
BUILDERARCH=$$(uname -m)

OUTPUTNAME=kdigger

RELEASE_NAME=kdigger-linux-amd64
RELEASE_FOLDER=release

# building for linux/amd64, if you want to build for arm64 you will have to
# adapt the syscall part that doesn't compile out of the box right now
GOOS=linux
GOARCH=amd64

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
build: lint
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags $(LDFLAGS) -o $(OUTPUTNAME)

fast-build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags $(LDFLAGS) -o $(OUTPUTNAME)

.PHONY: lint
lint:
	golangci-lint run

release: build
	mkdir $(RELEASE_FOLDER)
	mv kdigger $(RELEASE_FOLDER)/$(RELEASE_NAME)
	cd $(RELEASE_FOLDER) \
		&& sha256sum $(RELEASE_NAME) > $(RELEASE_NAME).sha256 \
		&& tar cvf - $(RELEASE_NAME) | gzip -9 - > $(RELEASE_NAME).tar.gz

DEV_IMAGE_TAG=mtardy/kdigger-dev
.PHONY: run
run: fast-build
	echo "FROM mtardy/koolbox\nCOPY kdigger /usr/local/bin/kdigger" | docker build -t $(DEV_IMAGE_TAG) -f- .
	kind load docker-image $(DEV_IMAGE_TAG)
	kubectl run kdigger-dev-tmp --rm -i --tty --image $(DEV_IMAGE_TAG) --image-pull-policy Never -- bash

.PHONY: clean-docker
clean-docker:
	docker image rm $(DEV_IMAGE_TAG)

.PHONY: clean
clean:
	rm kdigger

.PHONY: clean-all
clean-all: clean clean-docker
