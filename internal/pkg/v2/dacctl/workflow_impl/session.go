package workflow_impl

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/workflow"
	"math"
	"math/rand"
	"time"
)

func NewSessionWorkflow() workflow.Session {
	return sessionWorkflow{}
}

type sessionWorkflow struct {
	session registry.SessionRegistry
	actions  registry.SessionActions
	allocations registry.AllocationRegistry
	pool registry.PoolRegistry
}

func (s sessionWorkflow) CreateSessionVolume(session datamodel.Session) error {
	// TODO needs to get the allocation mutex, create the session, then create the allocations
	//   failing if the pool isn't known, or doesn't have enough space
	err := s.validateSession(session)
	if err != nil {
		return err
	}

	// Get session lock, drop on function exit
	sessionMutex, err := s.session.GetSessionMutex(session.Name)
	if err != nil {
		return fmt.Errorf("unable to get session mutex: %s due to: %s", session.Name, err)
	}
	err = sessionMutex.Lock(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to lock session mutex: %s due to: %s", session.Name, err)
	}

	// Allocate bricks, and choose brick host server
	session, err = s.doSessionAllocation(session)
	if err != nil {
		sessionMutex.Unlock(context.TODO())
		return err
	}

	// Create filesystem on the brick host server
	// TODO: add timeout
	eventChan, err := s.actions.CreateSessionVolume(context.TODO(), session.Name)
	if err != nil {
		return err
	}

	// Drop mutex so the server can take it
	err = sessionMutex.Unlock(context.TODO())
	if err != nil {
		// TODO: cancel above action?
		return err
	}

	// Wait for the server to create the filesystem
	sessionAction := <-eventChan
	return sessionAction.Error
}

func (s sessionWorkflow) validateSession(session datamodel.Session) error {
	_, err := s.pool.GetPool(session.VolumeRequest.PoolName)
	if err != nil {
		return fmt.Errorf("invalid session, unable to find pool %s", session.VolumeRequest.PoolName)
	}
	// TODO: check for multi-job restrictions, etc?
	return nil
}


func (s sessionWorkflow) doSessionAllocation(session datamodel.Session) (datamodel.Session, error) {
	allocationMutex, err := s.allocations.GetAllocationMutex()
	if err != nil {
		return session, err
	}

	err = allocationMutex.Lock(context.TODO())
	if err != nil {
		return session, err
	}
	defer allocationMutex.Unlock(context.TODO())

	// Write allocations first
	actualSizeBytes, allocations, err := s.getBricks(session.VolumeRequest.PoolName, session.VolumeRequest.TotalCapacityBytes)
	if err != nil {
		return session, fmt.Errorf("can't allocate for session: %s due to %s", session.Name, err)
	}
	session.ActualSizeBytes = actualSizeBytes
	s.allocations.CreateAllocations(session.Name, allocations)

	// Create initial version of session
	session, err = s.session.CreateSession(session)
	if err != nil {
		// TODO: remove allocations
		return session, err
	}
	return session, err
}

func (s sessionWorkflow) getBricks(poolName datamodel.PoolName, bytes int) (int, []datamodel.Brick, error) {
	pool, err := s.allocations.GetPoolInfo(poolName)
	if err != nil {
		return 0, nil, err
	}

	bricksRequired := int(math.Ceil(float64(bytes) / float64(pool.Pool.GranularityBytes)))
	actualSize := bricksRequired * int(pool.Pool.GranularityBytes)

	bricks := getBricks(bricksRequired, pool)
	if len(bricks) != bricksRequired {
		return 0, nil, fmt.Errorf(
			"unable to get number of requested bricks (%d) for given pool (%s)",
			bricksRequired, pool.Pool.Name)
	}
	return actualSize, bricks, nil
}

func getBricks(bricksRequired int, poolInfo datamodel.PoolInfo) []datamodel.Brick {
	// pick some of the available bricks
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator

	var chosenBricks []datamodel.Brick
	randomWalk := r.Perm(len(poolInfo.AvailableBricks))
	for _, i := range randomWalk {
		candidateBrick := poolInfo.AvailableBricks[i]

		// TODO: should not the random walk mean this isn't needed!
		goodCandidate := true
		for _, brick := range chosenBricks {
			if brick == candidateBrick {
				goodCandidate = false
				break
			}
		}
		if goodCandidate {
			chosenBricks = append(chosenBricks, candidateBrick)
		}
		if len(chosenBricks) >= bricksRequired {
			break
		}
	}
	return chosenBricks
}

func (s sessionWorkflow) DeleteSession(sessionName datamodel.SessionName, hurry bool) error {
	// TODO get the session mutex, then update the session, then send the event
	//   note the actions registry will error out if the host is not up
	//   release the mutex and wait for the server to complete its work or we timeout
	panic("implement me")
}

func (s sessionWorkflow) DataIn(sessionName datamodel.SessionName) error {
	panic("implement me")
}

func (s sessionWorkflow) AttachVolumes(sessionName datamodel.SessionName, computeNodes []string, loginNodes []string) error {
	panic("implement me")
}

func (s sessionWorkflow) DetachVolumes(sessionName datamodel.SessionName) error {
	panic("implement me")
}

func (s sessionWorkflow) DataOut(sessionName datamodel.SessionName) error {
	panic("implement me")
}

func (s sessionWorkflow) GetPools() ([]datamodel.PoolInfo, error) {
	panic("implement me")
}

func (s sessionWorkflow) GetSession(sessionName datamodel.SessionName) (datamodel.Session, error) {
	panic("implement me")
}

func (s sessionWorkflow) GetAllSessions() ([]datamodel.Session, error) {
	panic("implement me")
}
