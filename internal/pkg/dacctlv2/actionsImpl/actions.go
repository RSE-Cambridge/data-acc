package actionsImpl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/actions"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/parsers/capacity"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/session"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fileio"
	"log"
	"strings"
	"time"
)

var fakeTime uint = 0

func NewDacctlActions(registry session.Registry, actions session.Actions, disk fileio.Disk) actions.DacctlActions {
	return &dacctlActions{
		registry: registry,
		actions:  actions,
		disk:     disk,
		fakeTime: fakeTime,
	}
}

type dacctlActions struct {
	registry session.Registry
	actions  session.Actions
	disk     fileio.Disk
	fakeTime uint
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

var stringToAccessMode = map[string]datamodel.AccessMode{
	"":                datamodel.Striped,
	"striped":         datamodel.Striped,
	"private":         datamodel.Private,
	"private,striped": datamodel.PrivateAndStriped,
	"striped,private": datamodel.PrivateAndStriped,
}

func accessModeFromString(raw string) datamodel.AccessMode {
	return stringToAccessMode[strings.ToLower(raw)]
}

var stringToBufferType = map[string]datamodel.BufferType{
	"":        datamodel.Scratch,
	"scratch": datamodel.Scratch,
	"cache":   datamodel.Cache,
}

func bufferTypeFromString(raw string) datamodel.BufferType {
	return stringToBufferType[strings.ToLower(raw)]
}

func getNow() uint {
	return uint(time.Now().Unix())
}

func (d *dacctlActions) CreatePersistentBuffer(c actions.CliContext) error {
	checkRequiredStrings(c, "token", "caller", "capacity", "user", "access", "type")
	pool, capacityBytes, err := capacity.ParseCapacityBytes(c.String("capacity"))
	if err != nil {
		return err
	}
	request := datamodel.PersistentVolumeRequest{
		Caller:        c.String("caller"),
		CapacityBytes: capacityBytes,
		PoolName:      pool,
		Access:        accessModeFromString(c.String("access")),
		Type:          bufferTypeFromString(c.String("type")),
	}
	createdAt := d.fakeTime
	if createdAt == 0 {
		createdAt = getNow()
	}
	session := datamodel.Session{
		Name:                    datamodel.SessionName(c.String("token")),
		PersistentVolumeRequest: request,
		Owner:                   uint(c.Int("user")),
		Group:                   uint(c.Int("group")),
		CreatedAt:               createdAt,
	}
	session, err = d.registry.CreateSessionAllocations(session)
	if err != nil {
		return err
	}
	return d.actions.CreateSessionVolume(session)
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
