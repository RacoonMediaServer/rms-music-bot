package command

// Attributes means some traits about Command
type Attributes struct {
	// Internal - do not print command on help list
	Internal bool
	// CanRepeat - after command done, bot could pass next message to arguments
	CanRepeat bool
	// AuthRequired - need user credentials for run command
	AuthRequired bool
	// AdminRequired - flag admin rights is required to command
	AdminRequired bool
}

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

	// Attributes are traits of the command
	Attributes Attributes
}
