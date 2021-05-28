GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOFMT = $(GOCMD) fmt
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOVET = $(GOCMD) vet

BASE_NAME=imagingsvc

build: darwin web deploy-templates

all: darwin linux web deploy-templates

linux-full: linux web deploy-templates

darwin-full: darwin web deploy-templates

web:
	mkdir -p bin/
	cd frontend && yarn install && yarn build
	rm -rf bin/public
	mv frontend/dist bin/public

deploy-templates:
	mkdir -p bin/
	rm -rf bin/templates
	mkdir -p bin/templates
	cp ./templates/* bin/templates

darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -a -o bin/$(BASE_NAME).darwin backend/*.go

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o bin/$(BASE_NAME).linux backend/*.go

clean:
	rm -rf bin

fmt:
	cd backend; $(GOFMT)

vet:
	cd backend; $(GOVET)

dep:
	cd frontend && yarn upgrade
	$(GOGET) -u ./backend/...
	$(GOMOD) tidy
	$(GOMOD) verify
