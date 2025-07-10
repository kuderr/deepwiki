init:
	@echo "🔧 Initializing project dependencies..."
	@go mod tidy

format:
	@echo "🛠️ Formatting code..."
	@go tool golangci-lint fmt

pretty: format

check-format:
	@echo "🔍 Checking formatting..."
	@go tool golangci-lint fmt --diff > /dev/null || \
		( echo "❌ Found unformatted files. Run 'make format' to fix them."; exit 1 )

lint: check-format
	@echo "🚜 Linter goes brrrrrr..."
	@go tool golangci-lint run ./...

gogen:
	@echo "🔧 Generating code..."
	@go generate ./...

test:
	@echo "🏃 Running tests..."
	@CGO_ENABLED=1 go tool gotestsum -- --tags=test --race --vet= --count=10 ./...
