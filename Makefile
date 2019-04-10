GOCMD=$(shell which go)
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=$(shell basename `pwd`)
BINARY_UNIX=$(BINARY_NAME)_unix


all: test build

##build: build binary with go build 
build: 
	$(info building "$(BINARY_NAME)" binary)
	@$(GOBUILD) -o $(BINARY_NAME) -v

##test: test  go deps
test: 
	$(info Cheking dependencies)
	@$(GOTEST) -v ./... > /dev/null

##clean: exec go clean and delete binaries 
clean: 
	$(info  go clean and binary file deleting)
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_UNIX)

##install: copy binary file to /usr/local/bin dir  
install: 
	$(info copy "$(BINARY_NAME)" file to /usr/local/bin)
	@cp $(BINARY_NAME) /usr/local/bin

##deps: go get all required packages
deps:
	$(info go get: github.com/pymhd/go-logging, github.com/pymhd/go-simple-cache, github.com/mdirkse/i3ipc, github.com/docker/docker/client, github.com/docker/docker/api/types)
	@$(GOGET) github.com/pymhd/go-logging
	@$(GOGET) github.com/pymhd/go-simple-cache
	@$(GOGET) github.com/mdirkse/i3ipc
	@$(GOGET) github.com/docker/docker/client
	@$(GOGET) github.com/docker/docker/api/types

##help: show this message
help: Makefile
	@echo " Choose a command run in "$(PROJECTNAME)":"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'

