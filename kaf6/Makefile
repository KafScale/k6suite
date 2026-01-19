SHELL := /bin/sh

BINARY ?= ../kaf6-runner
SCENARIO ?= suite/smoke.json
KAF6_VERBOSE ?= 1
KAF6_DEBUG ?= 0
KAF6_OPEN ?= 1
.DEFAULT_GOAL := help

.PHONY: build run clean help run-s1-kcat run-suite run-select run-k6-select

build:
	go build -o $(BINARY) ./cmd/kaf6

run: build
	KAF6_VERBOSE=$(KAF6_VERBOSE) KAF6_DEBUG=$(KAF6_DEBUG) KAF6_OPEN=$(KAF6_OPEN) $(BINARY) run $(SCENARIO)

clean:
	rm -f $(BINARY)

help:
	@echo "Targets:"
	@echo "  build       - Build the kaf6 runner ($(BINARY))"
	@echo "  run         - Build and run the default scenario ($(SCENARIO))"
	@echo "  run-suite   - Run all kaf6 example scenarios"
	@echo "  run-select  - Select scenarios/profiles interactively"
	@echo "  run-k6-select - Select k6 scenarios/profiles interactively"
	@echo "  run-s1-kcat - Run Scenario 1 connectivity check with kcat"
	@echo "  clean       - Remove the built runner"
	@echo "  help        - Show this help text"

run-s1-kcat:
	sh ../scripts/kcat_s1_connectivity.sh

run-suite: build
	KAF6_VERBOSE=$(KAF6_VERBOSE) KAF6_DEBUG=$(KAF6_DEBUG) KAF6_OPEN=$(KAF6_OPEN) $(BINARY) run-suite suite

run-select: build
	KAF6_VERBOSE=$(KAF6_VERBOSE) KAF6_DEBUG=$(KAF6_DEBUG) KAF6_OPEN=$(KAF6_OPEN) $(BINARY) select suite

run-k6-select: build
	$(BINARY) k6-select ../tests/k6
