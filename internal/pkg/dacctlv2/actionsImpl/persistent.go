package actionsImpl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/actions"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/parsers/capacity"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"strings"
	"time"
)

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

var fakeTime uint = 0

func getNow() uint {
	if fakeTime != 0 {
		return fakeTime
	}
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
	session := datamodel.Session{
		Name:                    datamodel.SessionName(c.String("token")),
		PersistentVolumeRequest: request,
		Owner:                   uint(c.Int("user")),
		Group:                   uint(c.Int("group")),
		CreatedAt:               getNow(),
	}
	session, err = d.registry.CreateSessionAllocations(session)
	if err != nil {
		return err
	}
	return d.actions.CreateSessionVolume(session)
}