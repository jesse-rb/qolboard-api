RUN_GO = @docker compose run --service-ports --rm go

go-cli:
	@docker compose run --interactive --tty --rm go sh

adminer-up:
	docker-compose up adminer -d

adminer-down:
	docker-compose down adminer

db-up:
	docker-compose up db -d

db-down:
	docker-compose down db

docker-compose-destroy:
	docker-compose down -v

migrations-status:
	$(RUN_GO) go tool goose status

migrations-up:
	$(RUN_GO) go tool goose up

api-run:
	$(RUN_GO) go run main.go
