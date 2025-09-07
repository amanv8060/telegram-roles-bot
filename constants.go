package main

// Application constants
const (
	// Default configuration values
	DefaultHealthPort    = "8080"
	DefaultDatabasePath  = "bot.db"
	DefaultLogLevel      = "info"
	DefaultUpdateTimeout = 60
	DefaultMaxRetries    = 3
	DefaultRateLimit     = 30

	// Environment types
	EnvProduction  = "production"
	EnvDevelopment = "development"

	// Log levels
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"

	// Bot commands
	CmdPing           = "ping"
	CmdCreateRole     = "createrole"
	CmdRemoveRole     = "removerole"
	CmdAddToRole      = "addtorole"
	CmdRemoveFromRole = "removefromrole"
	CmdListRoles      = "listroles"
	CmdListMembers    = "listmembers"
	CmdHelp           = "help"
	CmdStatus         = "status"

	// Response messages
	MsgPong                = "🏓 pong"
	MsgUnauthorized        = "❌ You are not authorized to use this command."
	MsgProvideRoleName     = "❌ Please provide a role name."
	MsgUsageAddToRole      = "❌ Usage: /addtorole <rolename> <username>"
	MsgUsageRemoveFromRole = "❌ Usage: /removefromrole <rolename> <username>"
	MsgNoRoles             = "📋 No roles found."
	MsgBotHealthy          = "🟢 Bot is running and healthy!"
	MsgUnknownCommand      = "❌ Unknown command. Use /help to see available commands."

	// Response prefixes
	PrefixError   = "❌ Error: %v"
	PrefixSuccess = "✅ %s"
	PrefixInfo    = "📋 %s"
	PrefixPing    = "📢 Pinging role '%s': "

	// Help message template
	HelpMessage = `🤖 **Telegram Role Bot Commands**

**General Commands:**
/ping - Test if the bot is working
/ping <rolename> - Ping all users in a role
/listroles - List all roles
/listmembers <rolename> - List members of a role
/help - Show this help message

**Admin Commands:**
/createrole <rolename> - Create a new role
/removerole <rolename> - Remove a role
/addtorole <rolename> <username> - Add a user to a role
/removefromrole <rolename> <username> - Remove a user from a role

**Role Mentions:**
@<rolename> - Ping all users in a role

**Examples:**
/ping developers
/createrole developers
/addtorole developers john_doe
@developers`
)

// Admin commands that require special privileges
var AdminCommands = map[string]bool{
	CmdCreateRole:     true,
	CmdRemoveRole:     true,
	CmdAddToRole:      true,
	CmdRemoveFromRole: true,
}
