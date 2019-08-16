package datamodel

type SessionAction struct {
	Uuid    string
	Session Session
	Action  SessionActionType
	Error   error
}

type SessionActionType int

const (
	UnknownSessionAction SessionActionType = iota
	SessionDelete
	SessionCreate
	SessionCopyDataIn
	SessionMount
	SessionUnmount
	SessionCopyDataOut
)
