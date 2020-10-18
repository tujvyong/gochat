CURRENT_DIR := $(shell basename `pwd`)
GO_1 = $(CURRENT_DIR)_backend_1
GO_2 = $(CURRENT_DIR)_backend_2

up:
	@docker-compose up -d --scale backend=2

down:
	@docker-compose down

re: down up

redis:
	docker exec -it gochat_redis redis-cli

go1: up
	docker exec -it $(GO_1) sh
go2: up
	docker exec -it $(GO_2) sh

.PHONY: up down re redis go1 go2
