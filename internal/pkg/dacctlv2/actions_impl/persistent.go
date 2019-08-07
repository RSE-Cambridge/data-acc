package actions_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/actions"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctlv2/parsers/capacity"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/model"
	"strings"
	"time"
)

var stringToAccessMode = map[string]model.AccessMode{
	"":                model.Striped,
	"striped":         model.Striped,
	"private":         model.Private,
	"private,striped": model.PrivateAndStriped,
	"striped,private": model.PrivateAndStriped,
}

func accessModeFromString(raw string) model.AccessMode {
	return stringToAccessMode[strings.ToLower(raw)]
}

var stringToBufferType = map[string]model.BufferType{
	"":        model.Scratch,
	"scratch": model.Scratch,
	"cache":   model.Cache,
}

func bufferTypeFromString(raw string) model.BufferType {
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
	request := model.PersistentVolumeRequest{
		Caller:        c.String("caller"),
		CapacityBytes: capacityBytes,
		PoolName:      pool,
		Access:        accessModeFromString(c.String("access")),
		Type:          bufferTypeFromString(c.String("type")),
	}
	session := model.Session{
		Name:                    model.SessionName(c.String("token")),
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
