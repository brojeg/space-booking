# Simple Makefile for a Go project

# Build the application
all: build

build:
	@echo "Building..."
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go

# Create DB container
docker-run:
	@if docker compose up 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v

# Integration Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi


# Load .env file to access environment variables
include .env
export $(shell sed 's/=.*//' .env)

# Unified migration task with user prompt
migrate:
	@if command -v migrate > /dev/null; then \
            echo "Select migration option: [up/down/new]"; \
            read -p "Enter your choice: " choice; \
            if [ "$$choice" = "up" ]; then \
                migrate -path ./migrations -database "postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable&search_path=$(DB_SCHEMA)" up; \
            elif [ "$$choice" = "down" ]; then \
                migrate -path ./migrations -database "postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable&search_path=$(DB_SCHEMA)" down; \
            elif [ "$$choice" = "new" ]; then \
                read -p "Enter migration name: " name; \
                migrate create -dir ./migrations -ext sql $$name; \
            else \
                echo "Invalid choice. Please select 'up', 'down', or 'new'."; \
                exit 1; \
            fi; \
        else \
            read -p "'migrate' is not installed on your machine. Do you want to install it? [Y/n] " install_choice; \
            if [ "$$install_choice" != "n" ] && [ "$$install_choice" != "N" ]; then \
                go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; \
                echo "Select migration option: [up/down/new]"; \
                read -p "Enter your choice: " choice; \
                if [ "$$choice" = "up" ]; then \
                    migrate -path ./migrations -database "postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable&search_path=$(DB_SCHEMA)" up; \
                elif [ "$$choice" = "down" ]; then \
                    migrate -path ./migrations -database "postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_DATABASE)?sslmode=disable&search_path=$(DB_SCHEMA)" down; \
                elif [ "$$choice" = "new" ]; then \
                    read -p "Enter migration name: " name; \
                    migrate create -dir ./migrations -ext sql $$name; \
                else \
                    echo "Invalid choice. Please select 'up', 'down', or 'new'."; \
                    exit 1; \
                fi; \
            else \
                echo "You chose not to install 'migrate'. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run test clean watch docker-run docker-down migrate