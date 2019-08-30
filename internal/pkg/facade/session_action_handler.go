package facade

import "github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"

type SessionActionHandler interface {
	ProcessSessionAction(action datamodel.SessionAction)
}
