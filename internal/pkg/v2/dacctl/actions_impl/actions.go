package actions_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/workflow"
	"log"
	"strings"
)

func NewDacctlActions(registry registry.SessionRegistry, actions workflow.Session, disk fileio.Disk) dacctl.DacctlActions {
	return &dacctlActions{
		registry: registry,
		session:  actions,
		disk:     disk,
	}
}

type dacctlActions struct {
	registry registry.SessionRegistry
	session  workflow.Session
	disk     fileio.Disk
}

func checkRequiredStrings(c dacctl.CliContext, flags ...string) {
	var errs []string
	for _, flag := range flags {
		if str := c.String(flag); str == "" {
			errs = append(errs, flag)
		}
	}
	if len(errs) > 0 {
		log.Fatalf("Please provide these required parameters: %s", strings.Join(errs, ", "))
	}
}

func (d *dacctlActions) getSession(c dacctl.CliContext) (datamodel.Session, error) {
	// TODO - not sure we need this now...?
	checkRequiredStrings(c, "token")
	token := c.String("token")
	sessionName := datamodel.SessionName(token)
	s, err := d.registry.GetSession(sessionName)
	if err != nil {
		return s, fmt.Errorf("unable to find session for token %s", token)
	}
	return s, nil
}

func (d *dacctlActions) DeleteBuffer(c dacctl.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	return d.session.DeleteSession(s.Name)
}

func (d *dacctlActions) DataIn(c dacctl.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	return d.session.DataIn(s.Name)
}

func (d *dacctlActions) PreRun(c dacctl.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	// TODO - fix attach hosts
	return d.session.AttachVolumes(s.Name, nil)
}

func (d *dacctlActions) PostRun(c dacctl.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	return d.session.DetachVolumes(s.Name)
}

func (d *dacctlActions) DataOut(c dacctl.CliContext) error {
	s, err := d.getSession(c)
	if err != nil {
		return err
	}
	return d.session.DataOut(s.Name)
}
