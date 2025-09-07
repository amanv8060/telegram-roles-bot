# API Reference

## Bot Commands

### General Commands

#### `/ping`
Tests bot connectivity.
- **Usage**: `/ping`
- **Response**: "🏓 pong"
- **Access**: All users

#### `/ping <rolename>`
Pings all users in a specific role.
- **Usage**: `/ping developers`
- **Response**: "📢 Pinging role 'developers': @user1 @user2"
- **Access**: All users

#### `/listroles`
Lists all available roles.
- **Usage**: `/listroles`
- **Response**: "📋 Roles: developers, admins"
- **Access**: All users

#### `/listmembers <rolename>`
Lists all members of a specific role.
- **Usage**: `/listmembers developers`
- **Response**: "📋 Users in role 'developers': user1, user2"
- **Access**: All users

#### `/help`
Shows help message with all available commands.
- **Usage**: `/help`
- **Response**: Complete command reference
- **Access**: All users

#### `/status`
Shows bot health status.
- **Usage**: `/status`
- **Response**: "🟢 Bot is running and healthy!"
- **Access**: All users

### Admin Commands

#### `/createrole <rolename>`
Creates a new role.
- **Usage**: `/createrole developers`
- **Response**: "✅ Role 'developers' created successfully"
- **Access**: Admins only
- **Errors**: 
  - Role already exists
  - Invalid role name

#### `/removerole <rolename>`
Removes an existing role.
- **Usage**: `/removerole developers`
- **Response**: "✅ Role 'developers' removed successfully"
- **Access**: Admins only
- **Errors**: 
  - Role not found
  - Invalid role name

#### `/addtorole <rolename> <username>`
Adds a user to a role.
- **Usage**: `/addtorole developers john_doe`
- **Response**: "✅ User john_doe added to role 'developers'"
- **Access**: Admins only
- **Errors**: 
  - Role not found
  - Invalid username/role name

#### `/removefromrole <rolename> <username>`
Removes a user from a role.
- **Usage**: `/removefromrole developers john_doe`
- **Response**: "✅ User john_doe removed from role 'developers'"
- **Access**: Admins only
- **Errors**: 
  - Role not found
  - User not in role

### Role Mentions

#### `@<rolename>`
Alternative way to ping all users in a role.
- **Usage**: `@developers`
- **Response**: "📢 Pinging role @developers: @user1 @user2"
- **Access**: All users

## HTTP Endpoints

### Health Check

#### `GET /health`
Returns bot health status.
- **URL**: `http://localhost:8080/health`
- **Response**: 
  - `200 OK`: "HEALTHY"
  - `503 Service Unavailable`: "UNHEALTHY"

## Error Responses

### Format
All error responses follow this format:
```
❌ Error: <error_message>
```

### Common Errors

- **Unauthorized**: "❌ You are not authorized to use this command."
- **Invalid Input**: "❌ Error: invalid role name '': cannot be empty"
- **Not Found**: "❌ Error: role 'nonexistent' not found"
- **Already Exists**: "❌ Error: role 'developers' already exists"
- **Rate Limited**: Rate limit errors are handled silently

## Input Validation

### Role Names
- **Max Length**: 100 characters
- **Allowed Characters**: Alphanumeric, underscore, hyphen
- **Sanitization**: Removes newlines, carriage returns

### Usernames
- **Max Length**: 100 characters
- **Format**: Telegram username format
- **Sanitization**: Removes dangerous characters

### Message Length
- **Max Length**: 4000 characters (Telegram limit)
- **Validation**: Checked before processing

## Rate Limiting

- **Default**: 30 requests per minute per user
- **Configurable**: Via `RATE_LIMIT_PER_MIN` environment variable
- **Scope**: Per Telegram user ID
- **Response**: Silent rejection (no error message)
