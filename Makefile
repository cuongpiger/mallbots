tidy:
	go mod tidy

setup:
	docker compose -f docker-compose-dev.yml up -d

teardown:
	docker compose -f docker-compose-dev.yml down

run:
	export $$(cat ./docker/.env-dev | xargs) && go run ./cmd/mallbots/

generate:
	@echo running code generation
	@go generate ./...
	@echo done

.PHONY: tidy setup teardown run generate
