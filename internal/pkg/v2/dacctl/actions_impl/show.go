package actions_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
)

func (d *dacctlActions) getSession(c dacctl.CliContext) (datamodel.Session, error) {
	sessionName, err := d.getSessionName(c)
	if err != nil {
		return datamodel.Session{}, err
	}
	return d.session.GetSession(sessionName)
}

func (d *dacctlActions) RealSize(c dacctl.CliContext) (string, error) {
	session, err := d.getSession(c)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		`{"token":"%s", "capacity":%d, "units":"bytes"}`,
		session.Name, session.ActualSizeBytes), nil
}

func (d *dacctlActions) Paths(c dacctl.CliContext) error {
	err := checkRequiredStrings(c, "token", "pathfile")
	if err != nil {
		return err
	}

	session, err := d.getSession(c)
	if err != nil {
		return err
	}

	var paths []string
	for key, value := range session.Paths {
		paths = append(paths, fmt.Sprintf("%s=%s", key, value))
	}
	return d.disk.Write(c.String("pathfile"), paths)
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
