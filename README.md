# Telegram Role Bot

> Telegram bot for managing roles and pinging users in groups.

## Features

- **Role Management** - Create, remove, and manage roles
- **User Assignment** - Add/remove users from roles  
- **Role Pinging** - Ping all users in a role with `/ping rolename` or `@rolename`
- **Admin Controls** - Restrict sensitive operations to admin users
- **Security** - Rate limiting, input validation, and access controls
- **Monitoring** - Health checks and structured logging
- **Docker Ready** - Containerized deployment with Docker Compose

## Quick Start

### Prerequisites
- Go 1.22+ (for local development)
- Docker & Docker Compose (for deployment)
- Telegram Bot Token from [@BotFather](https://t.me/botfather)

### Local Development

```bash
# Clone repository
git clone <repository-url>
cd didactic-spork

# Setup environment
cp deployments/env.example .env
# Edit .env with your bot token and admin username

# Run the bot
go run cmd/bot/main.go
```

### Docker Deployment

```bash
# Copy environment file
cp deployments/env.example .env
# Edit .env with your configuration

# Deploy with Docker Compose
cd deployments
docker-compose up -d
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `TELEGRAM_APITOKEN` | Telegram bot token (required) | - |
| `ADMIN_USERNAME` | Admin username (required) | - |
| `DATABASE_PATH` | SQLite database file path | `bot.db` |
| `LOG_LEVEL` | Logging level (debug/info/warn/error) | `info` |
| `HEALTH_PORT` | Health check server port | `8080` |

## Commands

### General Commands
- `/ping` - Test bot connectivity
- `/ping <rolename>` - Ping all users in a role
- `/listroles` - List all available roles
- `/listmembers <rolename>` - List members of a role
- `/help` - Show help message

### Admin Commands
- `/createrole <rolename>` - Create a new role
- `/removerole <rolename>` - Remove a role
- `/addtorole <rolename> <username>` - Add user to role
- `/removefromrole <rolename> <username>` - Remove user from role

### Role Mentions
- `@<rolename>` - Ping all users in a role

## Project Structure

```
├── cmd/bot/              # Application entry point
├── internal/             # Private application code
│   ├── bot/             # Bot service and main logic
│   ├── config/          # Configuration management
│   ├── database/        # Database initialization
│   ├── handlers/        # Command handlers
│   ├── middleware/      # Security and rate limiting
│   ├── models/          # Data models and constants
│   └── store/           # Data storage operations
├── pkg/                 # Public library code
│   ├── logger/          # Logging utilities
│   └── utils/           # Common utilities
├── deployments/         # Deployment configurations
└── docs/               # Documentation
```

## Development

### Build
```bash
go build -o bin/bot cmd/bot/main.go
```

### Test
```bash
go test ./...
```

### Format
```bash
go fmt ./...
```

## Deployment

### Docker
```bash
cd deployments
docker build -t telegram-role-bot .
docker run --env-file .env -p 8080:8080 telegram-role-bot
```

### Docker Compose
```bash
cd deployments
docker-compose up -d
```

### Health Monitoring
```bash
curl http://localhost:8080/health
```

## Security

- **Rate Limiting** - Prevents spam and abuse
- **Input Validation** - Sanitizes all user inputs  
- **Admin Restrictions** - Sensitive operations require admin privileges
- **SQL Injection Protection** - Parameterized queries
- **Non-root Container** - Runs as unprivileged user

## License

MIT License