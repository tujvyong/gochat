up:
	@docker-compose up -d

down:
	docker-compose down

re: down up

redis:
	docker exec -it gochat_redis redis-cli

golang: up
	docker exec -it gochat_golang sh

.PHONY: up down re redis golang
