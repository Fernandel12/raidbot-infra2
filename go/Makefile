GO ?= go
GOPATH ?= $(HOME)/go
GO_TEST_PATH ?= ./...
GO_INSTALL_OPTS = -v -ldflags "-s -w"
GO_TEST_OPTS ?= -test.timeout=300s -race -cover -coverprofile=coverage.txt -covermode=atomic

check-program = $(foreach exec,$(1),$(if $(shell PATH="$(PATH)" which $(exec)),,$(error "No $(exec) in PATH")))

## MYSQL
MYSQL_CONFIG = -u root -puns3cur3
MYSQL_URN := root:uns3cur3@tcp(127.0.0.1:3306)/rslbot?charset=utf8mb4&parseTime=true

## API BIND
API_BIND ?= :8080

COMPILEDAEMON_OPTIONS ?= -exclude-dir=.git -color=true -build=go\ install -build-dir=./cmd/rslbot
COMPOSE_OPTS = -p rslbot -f docker-compose.yml -f docker-compose.dev.yml

## CLI OPTS
API_DEV_OPTS ?= --db-urn=$(MYSQL_URN) --bind=$(API_BIND)

## API RULES
.PHONY: api
api: mysql.up
api: redis.up
	GOFLAGS= go install github.com/githubnemo/CompileDaemon@latest
	$(GO) install $(GO_INSTALL_OPTS) ./cmd/rslbot
	CompileDaemon $(COMPILEDAEMON_OPTIONS) -command="rslbot api $(API_DEV_OPTS)"

.PHONY: api.down
api.down: mysql.down

## DOCKER RULES
.PHONY: up
up:
	docker-compose $(COMPOSE_OPTS) up -d

.PHONY: mysql.up
mysql.up:
	docker-compose $(COMPOSE_OPTS) up -d mysql

.PHONY: mysql.flush
mysql.flush: mysql.down
	docker volume rm -f raidbot_mysql_data

.PHONY: mysql.down
mysql.down:
	docker-compose -p rslbot stop mysql || true
	docker-compose -p rslbot rm -f -v mysql || true

.PHONY: redis.up
redis.up:
	docker-compose $(COMPOSE_OPTS) up -d redis

.PHONY: redis.down
redis.down:
	docker-compose -p rslbot stop redis || true
	docker-compose -p rslbot rm -f -v redis || true

.PHONY: mysql.logs
mysql.logs:
	docker-compose -p rslbot logs --tail=1000 -f mysql

.PHONY: mysql.shell
mysql.shell:
	docker-compose -p rslbot exec mysql mysql $(MYSQL_CONFIG) rslbot

## GENERAL RULES
.PHONY: test
test: unittest lint tidy

.PHONY: unittest
unittest: generate
	$(GO) test $(GO_TEST_OPTS) $(GO_TEST_PATH)

.PHONY: lint
lint: generate
	golangci-lint run --timeout=120s --verbose ./...

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: check-error-codes
check-error-codes:
	@./scripts/check-error-codes.sh

.PHONY: install
install: generate tidy
	$(GO) install $(GO_INSTALL_OPTS) ./cmd/...

.PHONY: clean
clean:
	rm -rf out/
	rm -rf gen/
	rm -f $(GEN_SUM)

PROTOS_SRC := $(wildcard ../api/proto/rslbot/*.proto)
GEN_SRC := $(PROTOS_SRC) ../api/buf.yaml ../api/buf.gen.yaml
GEN_SUM := gen.sum

.PHONY: generate
generate: $(GEN_SUM)

$(GEN_SUM): $(GEN_SRC)
	$(call check-program, shasum $(GO))
	@shasum $(GEN_SRC) | sort -k 2 > $(GEN_SUM).tmp
	@if [ ! -f $(GEN_SUM) ]; then \
		cd .. && make buf-generate; \
	else \
		diff -q $(GEN_SUM).tmp $(GEN_SUM) || ( \
			cd .. && make buf-generate \
		); \
	fi
	@mv $(GEN_SUM).tmp $(GEN_SUM)
