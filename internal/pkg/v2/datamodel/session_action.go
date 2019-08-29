package datamodel

type SessionAction struct {
	Uuid       string
	Session    Session
	ActionType SessionActionType
	Error      string
}

type SessionActionType string

// TODO: probably should be an int with custom parser?
const (
	UnknownSessionAction    SessionActionType = SessionActionType("")
	SessionCreateFilesystem                   = SessionActionType("CreateFilesystem")
	SessionDelete                             = SessionActionType("Delete")
	SessionCopyDataIn                         = SessionActionType("CopyDataIn")
	SessionMount                              = SessionActionType("Mount")
	SessionUnmount                            = SessionActionType("Unmount")
	SessionCopyDataOut                        = SessionActionType("CopyDataOut")
)
