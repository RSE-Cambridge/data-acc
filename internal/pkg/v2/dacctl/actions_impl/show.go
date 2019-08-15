package actions_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl"
)

func (d *dacctlActions) RealSize(c dacctl.CliContext) (string, error) {
	sessionName, err := d.getSessionName(c)
	if err != nil {
		return "", err
	}
	session, err := d.session.GetSession(sessionName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`{"token":"%s", "capacity":%d, "units":"bytes"}`,
		session.Name, session.ActualSizeBytes), nil
}

func (d *dacctlActions) Paths(c dacctl.CliContext) error {
	panic("implement me")
}

func (d *dacctlActions) ShowInstances() error {
	panic("implement me")
}

func (d *dacctlActions) ShowSessions() error {
	panic("implement me")
}

func (d *dacctlActions) ListPools() error {
	panic("implement me")
}

func (d *dacctlActions) ShowConfigurations() error {
	panic("implement me")
}
