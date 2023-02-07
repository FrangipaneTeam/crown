package slashcommand

import "errors"

//go:generate stringer -type=Command
type Command int

const (
	CommandLabel Command = 69 << iota
	CommandTrack
)

type Verb string

const (
	VerbAdd    = "add"
	VerbRemove = "remove"
)

// findVerb finds verb in command string.
func findVerb(cmd string) (Verb, error) {
	switch cmd {
	case VerbAdd:
		return VerbAdd, nil
	case VerbRemove:
		return VerbRemove, nil
	default:
		return "", errors.New("invalid verb")
	}
}

type SlashCommandLabel struct {
	Action Command
	Verb   Verb
	Label  string
}

type SlashCommandTrack struct {
	Action Command
}
