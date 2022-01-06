.PHONY: test-migrations
test-migrations:
	docker rm -f test-migrations ||:
	docker run -d --name test-migrations -e POSTGRES_DB=test-migrations -e POSTGRES_USER=test-migrations -e POSTGRES_PASSWORD=test-migrations -p5432:5432 postgres:13-bullseye
	docker exec -it test-migrations sh -c 'until pg_isready; do sleep 1; done; sleep 1'
	go run ./cmd/migrate '-db-url=postgres://test-migrations:test-migrations@localhost:5432/test-migrations?sslmode=disable'
	go run ./cmd/testreport '-db-url=postgres://test-migrations:test-migrations@localhost:5432/test-migrations?sslmode=disable'
	docker rm -f test-migrations
