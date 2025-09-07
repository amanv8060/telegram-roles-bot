# Architecture

## Overview

The Telegram Role Bot follows a clean architecture pattern with clear separation of concerns and dependency injection.

## Project Structure

```
cmd/bot/              # Application entry point
├── main.go          # Main function and application bootstrap

internal/            # Private application code (not importable)
├── bot/             # Bot service and orchestration
├── config/          # Configuration management
├── database/        # Database initialization and schema
├── handlers/        # Command handlers (business logic)
├── middleware/      # Security, rate limiting, validation
├── models/          # Data models, errors, and constants
└── store/           # Data persistence layer

pkg/                 # Public library code (importable)
├── logger/          # Structured logging utilities
└── utils/           # Common utility functions

deployments/         # Deployment configurations
├── Dockerfile       # Container definition
├── docker-compose.yml # Orchestration
├── Makefile         # Build automation
└── env.example      # Configuration template
```

## Design Principles

### 1. Clean Architecture
- **Separation of Concerns**: Each package has a single responsibility
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Interface Segregation**: Small, focused interfaces

### 2. Error Handling
- **Custom Error Types**: Structured errors with context
- **Error Wrapping**: Preserves error chains with `%w` verb
- **Graceful Degradation**: Non-critical errors don't crash the app

### 3. Security
- **Input Validation**: All user inputs are sanitized
- **Rate Limiting**: Per-user request throttling
- **Access Control**: Admin-only command restrictions
- **SQL Injection Protection**: Parameterized queries

### 4. Observability
- **Structured Logging**: JSON logs in production
- **Health Checks**: HTTP endpoints for monitoring
- **Error Context**: Rich error information for debugging

## Component Interaction

```
[Telegram API] -> [Bot Service] -> [Handlers] -> [Store] -> [Database]
                       |              |
                 [Middleware]    [Models]
                       |
                 [Security]
```

### Flow Description

1. **Bot Service** receives updates from Telegram API
2. **Middleware** validates and rate-limits requests
3. **Handlers** process commands and business logic
4. **Store** manages data persistence
5. **Database** provides data storage

## Configuration

Configuration is managed through environment variables with sensible defaults:

- **Development**: Loads from `.env` file
- **Production**: Uses environment variables directly
- **Validation**: Required fields are validated at startup

## Database Design

### Schema
- **roles**: Role definitions
- **users**: User information
- **role_users**: Many-to-many relationship

### Features
- **Foreign Key Constraints**: Data integrity
- **Indexes**: Performance optimization
- **Transactions**: Atomic operations
- **WAL Mode**: Better concurrency

## Security Model

### Authentication
- **Bot Token**: Validates against Telegram API
- **Admin Verification**: Username-based admin identification

### Authorization
- **Command Restrictions**: Admin-only operations
- **Chat Restrictions**: Optional chat allowlisting
- **Rate Limiting**: Per-user request throttling

### Input Validation
- **Sanitization**: Removes dangerous characters
- **Length Limits**: Prevents abuse
- **Type Validation**: Ensures correct data types

## Deployment

### Local Development
- Direct Go execution
- SQLite database
- File-based logging

### Production
- Docker containers
- Persistent volumes
- JSON logging
- Health monitoring
