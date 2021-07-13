# Injection variables
CMDPACKAGE=github.com/mtardy/kdigger/commands
GITCOMMIT=$$(git rev-parse HEAD)
GOVERSION=$$(go version | awk '{print $$3}')
ARCH=$$(uname -m)

OUTPUTNAME=kdigger

# -w disable DWARF generation
# -s disable symbol table
LDFLAGS="-s -w                               \
	-X $(CMDPACKAGE).VERSION=$(VERSION)      \
	-X $(CMDPACKAGE).GITCOMMIT=$(GITCOMMIT)  \
	-X $(CMDPACKAGE).GOVERSION=$(GOVERSION)  \
	-X $(CMDPACKAGE).ARCH=$(ARCH)"


build:
	go build -ldflags $(LDFLAGS) -o $(OUTPUTNAME)
