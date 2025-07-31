tidy:
	go mod tidy

setup:
	docker compose -f docker-compose-dev.yml up -d

teardown:
	docker compose -f docker-compose-dev.yml down

run:
	set -a && source ./docker/.env-dev && set +a && go run ./cmd/mallbots/

generate:
	@echo running code generation
	@go generate ./...
	@echo done

clean:
	@docker volume rm mallbots_pgdata

.PHONY: tidy setup teardown run generate clean
