build:
	@go build -o bin/goapi

run: build
	@./bin/goapi

test:
	@go test -v ../..
