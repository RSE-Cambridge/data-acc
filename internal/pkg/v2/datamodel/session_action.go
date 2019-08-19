package datamodel

type SessionAction struct {
	Uuid       string
	Session    Session
	ActionType SessionActionType
	Error      error
}

type SessionActionType int

const (
	UnknownSessionAction SessionActionType = iota
	SessionCreateFilesystem
	SessionDelete
	SessionCopyDataIn
	SessionMount
	SessionUnmount
	SessionCopyDataOut
)
