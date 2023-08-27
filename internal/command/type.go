package command

// Type describes commands of specific types
type Type struct {
	// ID is a command name
	ID string

	// Title is a human-readable command name
	Title string

	// Help is a description of command
	Help string

	// Factory can create commands of specific type
	Factory Factory

	// Internal means that the command is not visible via /help
	Internal bool
}
