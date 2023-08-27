package command

import "strings"

// Arguments is a list of command given arguments
type Arguments []string

func (a Arguments) String() string {
	result := ""
	for _, arg := range a {
		result += arg + " "
	}
	return strings.TrimSpace(result)
}

// ParseArguments splits text to arguments
func ParseArguments(text string) Arguments {
	return strings.Split(text, " ")
}
