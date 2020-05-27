package workflow_impl

import (
	"context"
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/filesystem"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/filesystem_impl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry_impl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/store"
	"log"
	"math"
	"math/rand"
	"time"
)

func NewSessionFacade(keystore store.Keystore) facade.Session {
	return sessionFacade{
		session:     registry_impl.NewSessionRegistry(keystore),
		actions:     registry_impl.NewSessionActionsRegistry(keystore),
		allocations: registry_impl.NewAllocationRegistry(keystore),
		ansible:     filesystem_impl.NewAnsible(),
	}
}

type sessionFacade struct {
	session     registry.SessionRegistry
	actions     registry.SessionActions
	allocations registry.AllocationRegistry
	ansible     filesystem.Ansible
}

func (s sessionFacade) submitJob(sessionName datamodel.SessionName, actionType datamodel.SessionActionType,
	getSession func() (datamodel.Session, error)) error {
	// 30 min timeout to acquire lock and send action
	ctxt, cancelFunc := context.WithTimeout(context.Background(), time.Minute*30)
	defer func() {
		cancelFunc()
	}()

	sessionMutex, err := s.session.GetSessionMutex(sessionName)
	if err != nil {
		return fmt.Errorf("unable to get session mutex: %s due to: %s", sessionName, err)
	}
	err = sessionMutex.Lock(ctxt)
	if err != nil {
		return fmt.Errorf("unable to lock session mutex: %s due to: %s", sessionName, err)
	}

	session, err := getSession()
	if err != nil {
		unlockErr := sessionMutex.Unlock(context.TODO())
		if unlockErr != nil {
			log.Println("failed to drop mutex", unlockErr)
		}
		return err
	}
	if session.Name == "" {
		// skip processing for this session
		// e.g. its a delete and we have already been deleted
		unlockErr := sessionMutex.Unlock(context.TODO())
		if unlockErr != nil {
			log.Println("failed to drop mutex", unlockErr)
		}
		return nil
	}

	// This will error out if the host is not currently up
	sessionActions, err := s.actions.SendSessionAction(ctxt, actionType, session)
	// Drop mutex regardless of if we had an error or not
	mutexErr := sessionMutex.Unlock(context.TODO())
	if err != nil {
		return err
	}
	if mutexErr != nil {
		return mutexErr
	}

	// ensure we get one value, and the channel is closed
	var finalResult *datamodel.SessionAction
	for action := range sessionActions {
		if finalResult != nil {
			log.Panicf("unexpected multiple actions")
		}
		finalResult = &action
	}
	if finalResult == nil {
		log.Panicf("failed to get response")
	}

	// report and errors in the server response
	if finalResult.Error != "" {
		return errors.New(finalResult.Error)
	}
	return nil
}

func (s sessionFacade) CreateSession(session datamodel.Session) error {
	err := s.validateSession(session)
	if err != nil {
		return err
	}

	return s.submitJob(session.Name, datamodel.SessionCreateFilesystem,
		func() (datamodel.Session, error) {
			// Allocate bricks, and choose brick host server
			session, err := s.doAllocationAndWriteSession(session)
			if err != nil {
				return session, err
			}
			if session.ActualSizeBytes == 0 && len(session.MultiJobAttachments) == 0 {
				// Skip creating an empty filesystem
				return datamodel.Session{}, nil
			}
			return session, nil
		})
}

func (s sessionFacade) validateSession(session datamodel.Session) error {
	_, err := s.allocations.GetPool(session.VolumeRequest.PoolName)
	if err != nil {
		return fmt.Errorf("invalid session, unable to find pool %s", session.VolumeRequest.PoolName)
	}
	// TODO: validate multi-job volumes exist
	// TODO: check for multi-job restrictions, etc?
	return nil
}

func (s sessionFacade) doAllocationAndWriteSession(session datamodel.Session) (datamodel.Session, error) {
	if session.VolumeRequest.TotalCapacityBytes > 0 {
		allocationMutex, err := s.allocations.GetAllocationMutex()
		if err != nil {
			return session, err
		}

		err = allocationMutex.Lock(context.TODO())
		if err != nil {
			return session, err
		}
		defer allocationMutex.Unlock(context.TODO())

		// Write allocations before creating the session
		actualSizeBytes, chosenBricks, err := s.getBricks(session.VolumeRequest.PoolName, session.VolumeRequest.TotalCapacityBytes)
		if err != nil {
			return session, fmt.Errorf("can't allocate for session: %s due to %s", session.Name, err)
		}

		session.ActualSizeBytes = actualSizeBytes
		session.AllocatedBricks = chosenBricks
		session.PrimaryBrickHost = chosenBricks[0].BrickHostName
	} else {
		// Pick a random alive host to be the PrimaryBrickHost anyway
		pools, err := s.allocations.GetAllPoolInfos()
		if err != nil {
			return session, err
		}
		if len(pools) == 0 {
			return session, fmt.Errorf("unable to find any pools")
		}
		// TODO: need to pick the default pool, but right now only one
		poolInfo := pools[0]
		bricks := pickBricks(1, poolInfo)
		session.PrimaryBrickHost = bricks[0].BrickHostName
	}

	// Store initial version of session
	// returned session will have updated revision info
	return s.session.CreateSession(session)
}

func (s sessionFacade) getBricks(poolName datamodel.PoolName, bytes int) (int, []datamodel.Brick, error) {
	pool, err := s.allocations.GetPoolInfo(poolName)
	if err != nil {
		return 0, nil, err
	}

	bricksRequired := int(math.Ceil(float64(bytes) / float64(pool.Pool.GranularityBytes)))
	actualSize := bricksRequired * int(pool.Pool.GranularityBytes)

	bricks := pickBricks(bricksRequired, pool)
	if len(bricks) != bricksRequired {
		return 0, nil, fmt.Errorf(
			"unable to get number of requested bricks (%d) for given pool (%s)",
			bricksRequired, pool.Pool.Name)
	}
	// TODO: we should add allocation map into the pool status
	//   so we spot conflicts if the locks failed for some reason
	return actualSize, bricks, nil
}

func pickBricks(bricksRequired int, poolInfo datamodel.PoolInfo) []datamodel.Brick {
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

func (s sessionFacade) DeleteSession(sessionName datamodel.SessionName, hurry bool) error {
	return s.submitJob(sessionName, datamodel.SessionDelete,
		func() (datamodel.Session, error) {
			session, err := s.session.GetSession(sessionName)
			if err != nil {
				log.Println("Unable to find session, skipping delete:", sessionName)
				return session, nil
			}

			if session.Status.DeleteRequested {
				// TODO: is there anything we can do about this?
				log.Println("Warning, delete already called")
			}

			// Record we want this deleted, in case host is not alive
			// can be deleted when it is next stated
			session.Status.DeleteRequested = true
			session.Status.DeleteSkipCopyDataOut = hurry
			return s.session.UpdateSession(session)
		})
}

func (s sessionFacade) CopyDataIn(sessionName datamodel.SessionName) error {
	return s.submitJob(sessionName, datamodel.SessionCopyDataIn,
		func() (datamodel.Session, error) {
			return s.session.GetSession(sessionName)
		})
}

func (s sessionFacade) Mount(sessionName datamodel.SessionName, computeNodes []string, loginNodes []string) error {
	return s.submitJob(sessionName, datamodel.SessionMount,
		func() (datamodel.Session, error) {
			session, err := s.session.GetSession(sessionName)
			if err != nil {
				log.Println("Unable to find session we want to mount:", sessionName)
				return session, err
			}

			// TODO: what about the login nodes? what do we want to do there?
			session.RequestedAttachHosts = computeNodes
			return s.session.UpdateSession(session)
		})
}

func (s sessionFacade) Unmount(sessionName datamodel.SessionName) error {
	return s.submitJob(sessionName, datamodel.SessionUnmount,
		func() (datamodel.Session, error) {
			return s.session.GetSession(sessionName)
		})
}

func (s sessionFacade) CopyDataOut(sessionName datamodel.SessionName) error {
	return s.submitJob(sessionName, datamodel.SessionCopyDataOut,
		func() (datamodel.Session, error) {
			return s.session.GetSession(sessionName)
		})
}

func (s sessionFacade) GetPools() ([]datamodel.PoolInfo, error) {
	return s.allocations.GetAllPoolInfos()
}

func (s sessionFacade) GetSession(sessionName datamodel.SessionName) (datamodel.Session, error) {
	return s.session.GetSession(sessionName)
}

func (s sessionFacade) GetAllSessions() ([]datamodel.Session, error) {
	return s.session.GetAllSessions()
}

func (s sessionFacade) GenerateAnsible(sessionName datamodel.SessionName) (string, error) {
	session, err := s.session.GetSession(sessionName)
	if err != nil {
		log.Println("Unable to find session we want to mount:", sessionName)
		return "", err
	}
	return s.ansible.CreateEnvironment(session)
}
