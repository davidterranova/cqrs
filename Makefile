include .env
export $(shell sed 's/=.*//' .env.docker)

TEST_TAGS ?= unit,integration
TEST_VERBOSE ?= 

.PHONY: test
test:
	docker compose down
	docker compose up db -d
	docker compose run migrate
	docker compose run --no-deps --build cqrs-test go test $(TEST_VERBOSE) ./... --cover -count=1 --tags=$(TEST_TAGS)
	docker compose down

-%:
	-@$(MAKE) $*
