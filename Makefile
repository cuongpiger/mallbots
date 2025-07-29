tidy:
	go mod tidy

setup:
	docker compose -f docker-compose-dev.yml up -d

teardown:
	docker compose -f docker-compose-dev.yml down

run:
	export $$(cat ./docker/.env-dev | xargs) && go run ./cmd/mallbots/

.PHONY: tidy setup teardown
