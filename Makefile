SHELL := /bin/sh
MAKEFLAGS += --no-print-directory
export GOCACHE := $(CURDIR)/.cache/go-build
export CGO_ENABLED := 1
export GOTOOLCHAIN := local

.PHONY: verify format-check vet test build browser frontend-test health-history
verify:
	go run ./tests/evidence
	go run ./tests/review-history
	$(MAKE) health-history
	$(MAKE) format-check
	go mod verify
	$(MAKE) vet
	$(MAKE) frontend-test
	$(MAKE) test
	$(MAKE) build
	$(MAKE) browser

health-history:
	go run ./tests/health-history

format-check:
	@test -z "$$(gofmt -l cmd internal tests)" || { gofmt -l cmd internal tests; exit 1; }

vet:
	go vet ./...

test:
	go test -race -count=1 -v ./...

frontend-test:
	node --test tests/browser/*.test.mjs

build:
	@mkdir -p bin
	go build -trimpath -o bin/model-relay ./cmd/model-relay

browser:
	npm ci --ignore-scripts --omit=optional --no-audit --no-fund --cache .cache/npm
	go run ./tests/browser
