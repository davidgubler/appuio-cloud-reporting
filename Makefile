.PHONY: ensure-prometheus
ensure-prometheus:
	go run ./util/ensure_prometheus

.PHONY: test
test: ensure-prometheus
	docker rm -f test-migrations ||:
	docker run -d --name test-migrations -e POSTGRES_DB=test-migrations -e POSTGRES_USER=test-migrations -e POSTGRES_PASSWORD=test-migrations -p65432:5432 postgres:13-bullseye
	docker exec -t test-migrations sh -c 'until pg_isready; do sleep 1; done; sleep 1'
	go run ./cmd/migrate '-db-url=postgres://test-migrations:test-migrations@localhost:65432/test-migrations?sslmode=disable'
	go run ./cmd/migrate -seed '-db-url=postgres://test-migrations:test-migrations@localhost:65432/test-migrations?sslmode=disable'
	go run ./cmd/migrate -seed '-db-url=postgres://test-migrations:test-migrations@localhost:65432/test-migrations?sslmode=disable'
	go test ./... -args '-db-url=postgres://test-migrations:test-migrations@localhost:65432/test-migrations?sslmode=disable'
	go run ./cmd/testreport '-db-url=postgres://test-migrations:test-migrations@localhost:65432/test-migrations?sslmode=disable'
	docker rm -f test-migrations
