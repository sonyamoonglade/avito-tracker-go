
run:
	docker-compose -f ./docker/docker-compose.dev.yaml --env-file ./.env up --build

stop:
	docker-compose -f ./docker/docker-compose.dev.yaml down --remove-orphans

build:
	go build -o ./bin/main ./cmd/main.go
local: build
	./bin/main
	
