package actions_impl

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacctl"
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

func (d *dacctlActions) ShowInstances() (string, error) {
	allSessions, err := d.session.GetAllSessions()
	if err != nil {
		return "", err
	}

	instances := instances{}
	for _, session := range allSessions {
		instances = append(instances, instance{
			Id:       string(session.Name),
			Capacity: instanceCapacity{Bytes: uint(session.ActualSizeBytes)},
			Links:    instanceLinks{string(session.Name)},
		})
	}
	return instancesToString(instances), nil
}

func (d *dacctlActions) ShowSessions() (string, error) {
	allSessions, err := d.session.GetAllSessions()
	if err != nil {
		return "", err
	}

	sessions := sessions{}
	for _, s := range allSessions {
		sessions = append(sessions, session{
			Id:      string(s.Name),
			Created: s.CreatedAt,
			Owner:   s.Owner,
			Token:   string(s.Name),
		})
	}
	return sessonsToString(sessions), nil
}

func (d *dacctlActions) ListPools() (string, error) {
	allPools, err := d.session.GetPools()
	if err != nil {
		return "", err
	}

	pools := pools{}
	for _, regPool := range allPools {
		free := len(regPool.AvailableBricks)
		quantity := free + len(regPool.AllocatedBricks)
		pools = append(pools, pool{
			Id:          string(regPool.Pool.Name),
			Units:       "bytes",
			Granularity: regPool.Pool.GranularityBytes,
			Quantity:    uint(quantity),
			Free:        uint(free),
		})
	}
	return getPoolsAsString(pools), nil
}

func (d *dacctlActions) ShowConfigurations() (string, error) {
	// NOTE: Slurm doesn't read any of the output, so we don't send anything
	return configurationToString(configurations{}), nil
}
