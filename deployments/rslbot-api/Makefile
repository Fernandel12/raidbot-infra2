.PHONY: up down logs restart backup restore

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

restart:
	docker compose restart

backup:
	./db_backup.sh

restore:
	@echo "Usage: docker compose exec mysql mysql -u rslbot -p rslbot < backup.sql"
