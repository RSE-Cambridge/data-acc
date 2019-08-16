package facade

import "github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"

type SessionActionHandler interface {
	ProcessSessionAction(action datamodel.SessionAction)
}
