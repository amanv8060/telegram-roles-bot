// Package models defines constants used throughout the application.
package models

// Bot commands
const (
	CmdPing           = "ping"
	CmdCreateRole     = "createrole"
	CmdRemoveRole     = "removerole"
	CmdAddToRole      = "addtorole"
	CmdRemoveFromRole = "removefromrole"
	CmdListRoles      = "listroles"
	CmdListMembers    = "listmembers"
	CmdHelp           = "help"
	CmdStatus         = "status"
)

// Response messages
const (
	MsgPong                = "pong"
	MsgUnauthorized        = "You are not authorized to use this command."
	MsgProvideRoleName     = "Please provide a role name."
	MsgUsageAddToRole      = "Usage: /addtorole <rolename> <username>"
	MsgUsageRemoveFromRole = "Usage: /removefromrole <rolename> <username>"
	MsgNoRoles             = "No roles found."
	MsgBotHealthy          = "Bot is running and healthy!"
	MsgUnknownCommand      = "Unknown command. Use /help to see available commands."
)

// Response prefixes
const (
	PrefixError   = "Error: %v"
	PrefixSuccess = "%s"
	PrefixInfo    = "%s"
	PrefixPing    = "Pinging role '%s': "
)

// Help message
const HelpMessage = `**Telegram Role Bot Commands**

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
@developers

**Note:** All role names and usernames are automatically converted to lowercase for consistency.`

// Admin commands that require special privileges
var AdminCommands = map[string]bool{
	CmdCreateRole:     true,
	CmdRemoveRole:     true,
	CmdAddToRole:      true,
	CmdRemoveFromRole: true,
}
