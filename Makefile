GOPATH ?= $(shell go env GOPATH)
GOBIN ?= $(GOPATH)/bin
GOOS ?=linux"
GOARCH ?=amd64

.PHONY: test
test:
	go test -v -cover -race ./...

.PHONY: postgres
postgres:
	docker rm -f notifi-postgres || true
	docker run --name notifi-postgres -e POSTGRES_PASSWORD=postgres -d --rm -p 5432:5432 docker.io/library/postgres:12
	docker exec -it notifi-postgres bash -c 'while ! pg_isready; do sleep 1; done;'

.PHONY: rabbitmq
rabbitmq:
	docker rm -f notifi-rabbitmq || true
	docker run --name notifi-rabbitmq -d --rm -p 15672:15672 -p 5672:5672 rabbitmq:3-management