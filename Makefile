GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOFMT = $(GOCMD) fmt
GOVET = $(GOCMD) vet

build: darwin 

all: darwin linux

darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/v4search.darwin cmd/*.go
	cp -r i18n/ bin/i18n

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/v4search.linux cmd/*.go
	cp -r i18n/ bin/i18n

clean:
	$(GOCLEAN) cmd/
	rm -rf bin

fmt:
	cd cmd; $(GOFMT)

vet:
	cd cmd; $(GOVET)

check:
	go get honnef.co/go/tools/cmd/staticcheck
	~/go/bin/staticcheck -checks all,-S1002,-ST1003 cmd/*.go
