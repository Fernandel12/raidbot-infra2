all: generate test
.PHONY: all

.PHONY: install
install: buf-generate go-install

.PHONY: buf-generate
buf-generate:
	cd api && buf generate
	mv go/pkg/proto/rslbot/rbapi* go/pkg/rbapi
	mv go/pkg/proto/rslbot/rbdb* go/pkg/rbdb
	mv go/pkg/proto/rslbot/errcode* go/pkg/errcode
	find go/pkg/rbdb -type f -name "*.go" -exec sed -i.bak '/rbdb "rslbot.com\/go\/pkg\/rbdb"/d;s/rbdb\.//' {} \;
	find go/pkg/rbdb -name "*.bak" -delete
	rm -rf go/pkg/proto
	touch go/gen.sum
.PHONY: go-install
go-install:
	cd go && make install

test:
	cd go && make test
.PHONY: test

clean:
	cd go && make clean
.PHONY: clean

tidy:
	cd go && make tidy
.PHONY: tidy

##
## docker.build
##
docker.push: docker.build
docker.push:
	docker push rslbot/api

.PHONY: docker.build
docker.build:
	docker buildx build --platform linux/x86_64 -t rslbot/api . -f Dockerfile

##
## Database Backups
##

BACKUP_BASE := /data/db_backups
TIMESTAMP := $(shell date +%Y%m%d_%H%M%S)

.PHONY: backup-help
backup-help:
	@echo "Database Backup Commands:"
	@echo "  make backup-rslbot-api     - Backup RSLBot API database"
	@echo "  make backup-discourse       - Backup Discourse database"
	@echo "  make backup-all             - Backup all databases and create compressed archive"

.PHONY: backup-all
backup-all: backup-rslbot-api
	@echo "Creating compressed archive..."
	@tar czf $(BACKUP_BASE)/all_backups_$(TIMESTAMP).tar.gz \
		-C $(BACKUP_BASE) \
		rslbot-api/rslbot-api_$(TIMESTAMP).sql
	@echo "✓ All backups compressed to $(BACKUP_BASE)/all_backups_$(TIMESTAMP).tar.gz"

.PHONY: backup-rslbot-api
backup-rslbot-api:
	@mkdir -p $(BACKUP_BASE)/rslbot-api
	@if [ ! -f deployments/rslbot-api/.env ]; then \
		echo "❌ Error: deployments/rslbot-api/.env file not found"; \
		echo "  Please create the .env file with MYSQL_PASSWORD=your_password"; \
		exit 1; \
	fi
	@cd deployments/rslbot-api && \
		MYSQL_PASSWORD=$$(grep '^MYSQL_PASSWORD=' .env | cut -d '=' -f2-) && \
		if [ -z "$$MYSQL_PASSWORD" ]; then \
			echo "❌ Error: MYSQL_PASSWORD not set in .env file"; \
			exit 1; \
		fi && \
		docker compose exec -T mysql mysqldump \
			-u rslbot \
			-p"$$MYSQL_PASSWORD" \
			--skip-lock-tables \
			--no-tablespaces \
			--routines \
			--triggers \
			--events \
			rslbot > $(BACKUP_BASE)/rslbot-api/rslbot-api_$(TIMESTAMP).sql
	@echo "✓ Backup saved to $(BACKUP_BASE)/rslbot-api/rslbot-api_$(TIMESTAMP).sql"
