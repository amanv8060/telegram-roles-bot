# Telegram Role Bot

A production-ready Telegram bot for managing roles and pinging users in groups. Built with Go and SQLite for reliability and performance.

## Features

- **Role Management**: Create, remove, and manage roles
- **User Assignment**: Add/remove users from roles
- **Role Pinging**: Ping all users in a role with `@rolename` or `/ping rolename`
- **Admin Controls**: Restrict sensitive operations to admin users
- **Persistent Storage**: SQLite database for data persistence
- **Rate Limiting**: Built-in rate limiting to prevent abuse
- **Health Checks**: HTTP endpoints for monitoring
- **Graceful Shutdown**: Proper cleanup on termination
- **Structured Logging**: JSON logging for production
- **Docker Support**: Ready for containerized deployment

## Quick Start

### Prerequisites

- Go 1.22+ (for local development)
- Docker (for containerized deployment)
- Telegram Bot Token (from [@BotFather](https://t.me/botfather))

### Local Development

1. **Clone and setup**:
   ```bash
   git clone <repository-url>
   cd didactic-spork
   make dev-setup
   ```

2. **Configure environment**:
   Edit `.env` file with your values:
   ```env
   TELEGRAM_APITOKEN=your_bot_token_here
   ADMIN_USERNAME=your_telegram_username
   ```

3. **Run the bot**:
   ```bash
   make run
   ```

### Docker Deployment

1. **Build and run**:
   ```bash
   make docker-build
   make docker-run
   ```

2. **Or use docker-compose**:
   ```bash
   make docker-compose-up
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `TELEGRAM_APITOKEN` | Telegram bot token (required) | - |
| `ADMIN_USERNAME` | Admin username (required) | - |
| `DATABASE_PATH` | SQLite database file path | `bot.db` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `ENV` | Environment (development, production) | `development` |
| `UPDATE_TIMEOUT` | Telegram update timeout in seconds | `60` |
| `RATE_LIMIT_PER_MIN` | Rate limit per user per minute | `30` |
| `HEALTH_PORT` | Health check server port | `8080` |
| `ALLOWED_CHATS` | Comma-separated list of allowed chat IDs (optional) | - |

### Example Configuration

```env
# Production configuration
TELEGRAM_APITOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
ADMIN_USERNAME=your_username
DATABASE_PATH=/app/data/bot.db
LOG_LEVEL=info
ENV=production
UPDATE_TIMEOUT=60
RATE_LIMIT_PER_MIN=30
HEALTH_PORT=8080
ALLOWED_CHATS=123456789,-987654321
```

## Commands

### General Commands
- `/ping` - Test if the bot is working
- `/ping <rolename>` - Ping all users in a role
- `/listroles` - List all roles
- `/listmembers <rolename>` - List members of a role
- `/help` - Show help message
- `/status` - Check bot status

### Admin Commands
- `/createrole <rolename>` - Create a new role
- `/removerole <rolename>` - Remove a role
- `/addtorole <rolename> <username>` - Add a user to a role
- `/removefromrole <rolename> <username>` - Remove a user from a role

### Role Mentions
- `@<rolename>` - Ping all users in a role

## Health Monitoring

The bot exposes HTTP endpoints for health monitoring:

- `GET /health` - Health check endpoint
- `GET /ready` - Readiness check endpoint

Example:
```bash
curl http://localhost:8080/health
```

## Database Schema

The bot uses SQLite with the following schema:

```sql
-- Roles table
CREATE TABLE roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    telegram_id INTEGER UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Role-User relationships
CREATE TABLE role_users (
    role_id INTEGER,
    user_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY(role_id, user_id)
);
```

## Security Features

- **Rate Limiting**: Prevents spam and abuse
- **Input Validation**: Sanitizes all user inputs
- **Admin Restrictions**: Sensitive operations require admin privileges
- **Chat Restrictions**: Optional restriction to specific chats
- **SQL Injection Protection**: Parameterized queries
- **Non-root Container**: Runs as non-privileged user in Docker

## Development

### Project Structure

```
.
├── main.go           # Main application entry point
├── config.go         # Configuration management
├── logger.go         # Logging setup
├── database.go       # Database initialization
├── store.go          # Data access layer
├── middleware.go     # Security and rate limiting
├── health.go         # Health check endpoints
├── Dockerfile        # Container definition
├── docker-compose.yml # Docker Compose configuration
├── Makefile          # Build automation
└── README.md         # This file
```

### Building

```bash
# Build for local development
make build

# Build Docker image
make docker-build

# Run tests
make test
```

### Adding New Features

1. Add new command handlers in `main.go`
2. Add database operations in `store.go` if needed
3. Update help text and documentation
4. Add tests for new functionality

## Deployment

### Docker Compose (Recommended)

```bash
# Start the bot
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the bot
docker-compose down
```

### Manual Docker

```bash
# Build image
docker build -t telegram-role-bot .

# Run container
docker run -d \
  --name telegram-bot \
  --env-file .env \
  -p 8080:8080 \
  -v bot_data:/app/data \
  telegram-role-bot
```

### Production Considerations

1. **Environment Variables**: Use proper secrets management
2. **Database Backups**: Regular backups of SQLite database
3. **Monitoring**: Set up monitoring for health endpoints
4. **Logging**: Configure log aggregation
5. **Updates**: Plan for zero-downtime updates

## Troubleshooting

### Common Issues

1. **Bot not responding**:
   - Check if token is correct
   - Verify bot is added to the group
   - Check logs for errors

2. **Database errors**:
   - Ensure database file is writable
   - Check disk space
   - Verify database permissions

3. **Rate limiting**:
   - Adjust `RATE_LIMIT_PER_MIN` if needed
   - Check for spam or abuse

### Logs

The bot provides structured logging. In production, logs are in JSON format:

```json
{
  "level": "info",
  "msg": "Role created successfully",
  "role": "developers",
  "time": "2024-01-01T12:00:00Z"
}
```

## License

This project is licensed under the MIT License.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For issues and questions:
1. Check the troubleshooting section
2. Review the logs
3. Open an issue on GitHub