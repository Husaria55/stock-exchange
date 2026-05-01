.PHONY: up down test chaos db-logs

up:
	docker compose up --build -d

down:
	docker compose down -v

test:
	RUN_INTEGRATION_TESTS=1 go test -v -race ./...

chaos:
	curl -X POST http://localhost:8000/chaos

logs:
	docker compose logs -f api