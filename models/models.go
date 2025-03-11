package models

// CommandConfig represents the complete command configuration
type CommandConfig struct {
	UE  NodeCommands `json:"ue"`
	GNB NodeCommands `json:"gnb"`
}

// NodeCommands represents a set of commands for a node type
type NodeCommands struct {
	Commands []Command `json:"commands"`
}

// Command represents a single command definition
type Command struct {
	Name         string       `json:"name"`
	Help         string       `json:"help"`
	Subcommands  []Subcommand `json:"subcommands"`
	DefaultUsage string       `json:"defaultUsage"`
}

// Subcommand represents a subcommand within a command
type Subcommand struct {
	Name     string `json:"name"`
	Help     string `json:"help"`
	Response string `json:"response"`
}
