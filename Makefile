# Makefile – команды для разработки Ricochet Task (CLI на Go)

BINARY := ricochet-task

.PHONY: build test lint docker clean mcp

build:
	go build -v -o $(BINARY) .

test:
	go test ./...

lint:
	go vet ./...

docker:
	docker build -t ricochet-task:latest -f Dockerfile ..

clean:
	rm -f $(BINARY)

# Запуск локального MCP-сервера
mcp:
	go run . mcp --port 8090 