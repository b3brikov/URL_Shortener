include .env

run:
	@docker compose up -d

down:
	@docker compose down -v

migrate-create:
	@docker compose run --rm migrations \
		create \
		-ext sql \
		-dir /migrations \
		-seq "${seq}"

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@docker compose run --rm migrations \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable \
		${action}


