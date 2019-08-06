package actionsImpl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/actions"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/session"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
)

func NewDacctlActions(registry session.Registry, actions session.Actions, disk fileio.Disk) actions.DacctlActions {
	return &dacctlActions{
		registry: registry,
		actions: actions,
		disk: disk,
	}
}

type dacctlActions struct {
	registry session.Registry
	actions session.Actions
	disk    fileio.Disk
}

func (d *dacctlActions) CreatePersistentBuffer(c actions.CliContext) error {
	bufferName := c.String("token")
	d.registry.GetSession(bufferName)
	return nil
}

func (*dacctlActions) DeleteBuffer(c actions.CliContext) error {
	panic("implement me")
}

func (*dacctlActions) CreatePerJobBuffer(c actions.CliContext) error {
	panic("implement me")
}

func (*dacctlActions) ShowInstances() error {
	panic("implement me")
}

func (*dacctlActions) ShowSessions() error {
	panic("implement me")
}

func (*dacctlActions) ListPools() error {
	panic("implement me")
}

func (*dacctlActions) ShowConfigurations() error {
	panic("implement me")
}

func (*dacctlActions) ValidateJob(c actions.CliContext) error {
	panic("implement me")
}

func (*dacctlActions) RealSize(c actions.CliContext) error {
	panic("implement me")
}

func (*dacctlActions) DataIn(c actions.CliContext) error {
	panic("implement me")
}

func (*dacctlActions) Paths(c actions.CliContext) error {
	panic("implement me")
}

func (*dacctlActions) PreRun(c actions.CliContext) error {
	panic("implement me")
}

func (*dacctlActions) PostRun(c actions.CliContext) error {
	panic("implement me")
}

func (*dacctlActions) DataOut(c actions.CliContext) error {
	panic("implement me")
}

