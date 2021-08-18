SHELL := /bin/bash

# The name of the executable (default is current directory name)
TARGET := $(shell echo $${PWD\#\#*/})
.DEFAULT_GOAL: $(TARGET)

# These will be provided to the target
VERSION := 1.0.0
BUILD := `git rev-parse HEAD`

# Use linker flags to provide version/build settings to the target
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"

# go source files, ignore vendor directory
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

all: install

$(TARGET): $(SRC)
	@go build $(LDFLAGS) -o $(TARGET)

build: $(TARGET)
	@true

install:
	@go install $(LDFLAGS)

fmt:
	@gofmt -l -w $(SRC)

simplify:
	@gofmt -s -l -w $(SRC)

test:
	@docker-compose -f docker-compose.test.yaml up --build --abort-on-container-exit
	@docker-compose -f docker-compose.test.yaml down --volumes