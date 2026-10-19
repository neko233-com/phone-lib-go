# Makefile for phone-lib-go.
#
# Requires GNU make and the Go toolchain; the recipes are plain "go" commands,
# so they work with both POSIX shells and the Windows cmd.exe shell.

GO ?= go

# Append .exe on Windows so that the binary is directly runnable.
EXE :=
ifeq ($(OS),Windows_NT)
EXE := .exe
endif

BINARY  ?= phone$(EXE)
NUMBER  ?= 13128500205
COUNTRY ?= CN

.PHONY: all build test race vet fmt run-demo help

all: build

## build: compile the phone CLI to ./phone (./phone.exe on Windows).
build:
	$(GO) build -o $(BINARY) ./cmd/phone

## test: run the unit tests of every package.
test:
	$(GO) test ./...

## race: run the unit tests with the race detector (requires a C compiler;
##        GitHub Actions CI runs this on Linux).
race:
	$(GO) test -race ./...

## vet: run go vet on every package.
vet:
	$(GO) vet ./...

## fmt: format every Go source file.
fmt:
	$(GO) fmt ./...

## run-demo: run the CLI on a mainland China mobile number, pretty and JSON.
run-demo:
	$(GO) run ./cmd/phone --country $(COUNTRY) $(NUMBER)
	$(GO) run ./cmd/phone --json --country $(COUNTRY) $(NUMBER)

## help: list the available targets.
help:
	@echo "make build     build the phone CLI into ./$(BINARY)"
	@echo "make test      go test ./..."
	@echo "make race      go test -race ./..."
	@echo "make vet       go vet ./..."
	@echo "make fmt       go fmt ./..."
	@echo "make run-demo  look up $(NUMBER) with country $(COUNTRY)"
