package actions_impl

import (
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/workflow"
	"log"
	"regexp"
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

func checkRequiredStrings(c dacctl.CliContext, flags ...string) error {
	var errs []string
	for _, flag := range flags {
		if str := c.String(flag); str == "" {
			errs = append(errs, flag)
		}
	}
	if len(errs) > 0 {
		errStr := fmt.Sprintf("Please provide these required parameters: %s", strings.Join(errs, ", "))
		log.Println(errStr)
		return errors.New(errStr)
	}
	return nil
}

func (d *dacctlActions) getSessionName(c dacctl.CliContext) (datamodel.SessionName, error) {
	err := checkRequiredStrings(c, "token")
	if err != nil {
		return "", err
	}

	token := c.String("token")
	re := regexp.MustCompile("[a-zA-Z0-9]*")
	if !re.Match([]byte(`seafood fool`)) {
		return "", fmt.Errorf("badly formatted session name: %s", token)
	}

	return datamodel.SessionName(token), nil
}

func (d *dacctlActions) DeleteBuffer(c dacctl.CliContext) error {
	sessionName, err := d.getSessionName(c)
	if err != nil {
		return err
	}
	hurry := c.Bool("hurry")
	return d.session.DeleteSession(sessionName, hurry)
}

func (d *dacctlActions) DataIn(c dacctl.CliContext) error {
	sessionName, err := d.getSessionName(c)
	if err != nil {
		return err
	}
	return d.session.DataIn(sessionName)
}

func (d *dacctlActions) PreRun(c dacctl.CliContext) error {
	sessionName, err := d.getSessionName(c)
	if err != nil {
		return err
	}

	// TODO - fix hosts
	return d.session.AttachVolumes(sessionName, nil, nil)
}

func (d *dacctlActions) PostRun(c dacctl.CliContext) error {
	sessionName, err := d.getSessionName(c)
	if err != nil {
		return err
	}
	return d.session.DetachVolumes(sessionName)
}

func (d *dacctlActions) DataOut(c dacctl.CliContext) error {
	sessionName, err := d.getSessionName(c)
	if err != nil {
		return err
	}
	return d.session.DataOut(sessionName)
}
