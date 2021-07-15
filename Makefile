CMDPACKAGE=github.com/mtardy/kdigger/commands
VERSION=$$(git describe --tags 2>/dev/null || echo dev)
GITCOMMIT=$$(git rev-parse HEAD)
GOVERSION=$$(go version | awk '{print $$3}')
ARCH=$$(uname -m)

OUTPUTNAME=kdigger

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
	-X $(CMDPACKAGE).GOVERSION=$(GOVERSION)  \
	-X $(CMDPACKAGE).ARCH=$(ARCH)"

build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags $(LDFLAGS) -o $(OUTPUTNAME)
