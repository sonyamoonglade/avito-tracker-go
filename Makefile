
run:
	docker-compose -f ./docker/docker-compose.dev.yaml -p parser --env-file ./.env up --build

stop:
	docker-compose -f ./docker/docker-compose.dev.yaml down

build:
	go build -o ./bin/main ./cmd/main.go

local: build
	./bin/main
	
migrate-local-up:
	docker-compose -f docker/docker-compose.migrations.yml --env-file .env up
migrate-local-down:
	docker-compose -f docker/docker-compose.migrations.down.yml --env-file .env up
