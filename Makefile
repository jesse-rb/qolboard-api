local-adminer-up:
	docker-compose up adminer -d

local-adminer-down:
	docker-compose down adminer

local-db-up:
	docker-compose up db -d

local-db-down:
	docker-compose down db

local-destroy:
	docker-compose down -v

migrations-up:
	goose up

local-api-up:
	go run main.go
