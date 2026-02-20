.PHONY: all build run test proto migrate clean emulator-up emulator-down

BINARY_NAME=catalog-server
SPANNER_EMULATOR_HOST=localhost:9010
PROJECT_ID=test-project
INSTANCE_ID=test-instance
DATABASE_ID=test-database

all: proto build

build:
	go build -o bin/$(BINARY_NAME) ./cmd/server

run: build
	SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) \
	SPANNER_PROJECT=$(PROJECT_ID) \
	SPANNER_INSTANCE=$(INSTANCE_ID) \
	SPANNER_DATABASE=$(DATABASE_ID) \
	./bin/$(BINARY_NAME)

test:
	SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) \
	go test -v -count=1 ./tests/e2e/... ./internal/app/product/domain/...

proto:
	protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/product/v1/product_service.proto

migrate:
	@echo "Run migrations against Spanner emulator"
	@echo "Using: $(SPANNER_EMULATOR_HOST)"
	@echo "Migrations are applied automatically in test setup"
	@echo "For manual apply, use the gcloud CLI or admin API"

emulator-up:
	docker-compose up -d

emulator-down:
	docker-compose down

clean:
	rm -rf bin/
	go clean

tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
