package actions_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/actions"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/model"
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

func (d *dacctlActions) getSession(c actions.CliContext) (model.Session, error) {
	checkRequiredStrings(c, "token")
	token := c.String("token")
	s, err := d.registry.GetSession(token)
	if err != nil {
		return s, fmt.Errorf("unable to find session for token %s", token)
	}
	return s, nil
}

func (d *dacctlActions) DeleteBuffer(c actions.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	return d.actions.DeleteSession(s)
}

func (d *dacctlActions) DataIn(c actions.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	return d.actions.DataIn(s)
}

func (d *dacctlActions) PreRun(c actions.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	return d.actions.AttachVolumes(s)
}

func (d *dacctlActions) PostRun(c actions.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	return d.actions.DetachVolumes(s)
}

func (d *dacctlActions) DataOut(c actions.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	return d.actions.DataOut(s)
}
