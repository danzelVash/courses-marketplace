
.PHONY: bin-debs
bin-debs:
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: infra
infra:
	docker-compose up

.PHONY: migrations-up
migrations-up:
	goose -dir=schema/migrations --allow-missing postgres "host=localhost port=5432 user=danzelVash password=22332223 dbname=courses sslmode=disable" up

.PHONY: migrations-down
migrations-down:
	goose -dir=schema/migrations --allow-missing postgres "host=localhost port=5432 user=danzelVash password=22332223 dbname=courses sslmode=disable" down


.PHONY: infra-stop
infra-stop:
	docker-compose down