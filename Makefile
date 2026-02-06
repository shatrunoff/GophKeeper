VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"
TEST_DB_DSN=postgres://test:test@localhost:5433/gopherpass_test

libs:
	go mod tidy
	go mod vendor

build:
	go build -o bin/server ./cmd/server
	go build $(LDFLAGS) -o bin/gopherpass ./cmd/client

test-up:
	@docker compose -f docker-compose.test.yml up -d --wait

test-migrate:
	@docker compose -f docker-compose.test.yml exec -T postgres psql -U test -d gopherpass_test -f - < migrations/001_init.up.sql

test-down:
	@docker compose -f docker-compose.test.yml down -v

test: libs test-up test-migrate
	@TEST_DB_DSN=$(TEST_DB_DSN) go test ./... -cover || ($(MAKE) test-down && exit 1)
	@$(MAKE) test-down

version:
	./bin/gopherpass version