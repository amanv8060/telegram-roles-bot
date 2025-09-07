.PHONY: build run test clean docker-build docker-run help

# Default target
help:
	@echo "Available targets:"
	@echo "  build        - Build the Go application"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  up           - Start with docker-compose"
	@echo "  down         - Stop docker-compose"
	@echo "  logs         - View docker-compose logs"

# Build the application
build:
	cd .. && go build -o bin/bot cmd/bot/main.go

# Run the application
run:
	cd .. && go run cmd/bot/main.go

# Run tests
test:
	cd .. && go test -v ./...

# Clean build artifacts
clean:
	cd .. && rm -rf bin/ *.db

# Docker operations
docker-build:
	docker build -t telegram-role-bot ..

docker-run:
	docker run --env-file .env -p 8080:8080 -v $(PWD)/data:/app/data telegram-role-bot

# Docker Compose operations
up:
	docker-compose up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then cp env.example .env; echo "Created .env file from template. Please edit it with your values."; fi
	@echo "Development setup complete!"

# Production deployment
deploy:
	@echo "Deploying to production..."
	docker-compose -f docker-compose.yml up -d
	@echo "Deployment complete!"