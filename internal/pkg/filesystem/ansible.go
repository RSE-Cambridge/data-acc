package filesystem

import "github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"

type Ansible interface {
	// returns temp dir environment was created in
	CreateEnvironment(session datamodel.Session) (string, error)
}
