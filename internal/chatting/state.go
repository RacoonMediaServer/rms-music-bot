package chatting

type chatState int

const (
	stateUnknown chatState = iota
	stateNoAccess
	stateAccessRequested
	stateAccessGranted
	stateAccessDenied
)
