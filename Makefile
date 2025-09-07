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
	@echo "  docker-compose-up - Start with docker-compose"
	@echo "  docker-compose-down - Stop docker-compose"

# Build the application
build:
	go build -o bin/telegram-bot .

# Run the application
run:
	go run .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f bot.db

# Docker operations
docker-build:
	docker build -t telegram-role-bot .

docker-run:
	docker run --env-file .env -p 8080:8080 -v $(PWD)/data:/app/data telegram-role-bot

# Docker Compose operations
docker-compose-up:
	docker-compose up -d

docker-compose-down:
	docker-compose down

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
