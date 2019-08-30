package actions_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacctl"
	parsers2 "github.com/RSE-Cambridge/data-acc/internal/pkg/dacctl/actions_impl/parsers"
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

func (d *dacctlActions) CreatePersistentBuffer(c dacctl.CliContext) error {
	err := checkRequiredStrings(c, "token", "caller", "capacity", "access", "type")
	if err != nil {
		return err
	}
	pool, capacityBytes, err := parsers2.ParseCapacityBytes(c.String("capacity"))
	if err != nil {
		return err
	}
	request := datamodel.VolumeRequest{
		MultiJob:           true,
		Caller:             c.String("caller"),
		TotalCapacityBytes: capacityBytes,
		PoolName:           datamodel.PoolName(pool),
		Access:             accessModeFromString(c.String("access")),
		Type:               bufferTypeFromString(c.String("type")),
	}
	session := datamodel.Session{
		Name:          datamodel.SessionName(c.String("token")),
		VolumeRequest: request,
		Owner:         uint(c.Int("user")),
		Group:         uint(c.Int("group")),
		CreatedAt:     getNow(),
	}
	return d.session.CreateSession(session)
}
