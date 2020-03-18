# Makefile for Pcounter Store.

GOCMD=go
GOBUILD=$(GOCMD) build 
LDGLAGS="-X main.sha1ver=`git rev-parse HEAD` -X main.buildTime=`date +%Y-%m-%d_%T` -X main.gitver=`git describe --tag`"
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
RM=rm


# Main target

.PHONY: build 
build: pcounter-store

pcounter-store: pcounter-store.go
	$(GOBUILD) -ldflags $(LDFLAGS)

.PHONY: clean
clean: 
	$(RM) pcounter




