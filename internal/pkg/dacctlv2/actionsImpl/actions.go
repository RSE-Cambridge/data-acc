package actionsImpl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/actions"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/session"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"log"
	"strings"
)

func NewDacctlActions(registry session.Registry, actions session.Actions, disk fileio.Disk) actions.DacctlActions {
	return &dacctlActions{
		registry: registry,
		actions:  actions,
		disk:     disk,
	}
}

type dacctlActions struct {
	registry session.Registry
	actions  session.Actions
	disk     fileio.Disk
}

func checkRequiredStrings(c actions.CliContext, flags ...string) {
	errs := []string{}
	for _, flag := range flags {
		if str := c.String(flag); str == "" {
			errs = append(errs, flag)
		}
	}
	if len(errs) > 0 {
		log.Fatalf("Please provide these required parameters: %s", strings.Join(errs, ", "))
	}
}

func (d *dacctlActions) DeleteBuffer(c actions.CliContext) error {
	checkRequiredStrings(c, "token")
	token := c.String("token")
	session, err := d.registry.GetSession(token)
	if err != nil {
		return fmt.Errorf("unable to find session for token %s", token)
	}
	return d.actions.DeleteSession(session)
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
