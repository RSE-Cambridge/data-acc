package workflow_impl

import (
	"context"
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_store"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionFacade_CreateSession_NoBricks(t *testing.T) {
	initialSession := datamodel.Session{
		Name: "foo",
		VolumeRequest: datamodel.VolumeRequest{
			PoolName:           datamodel.PoolName("pool1"),
			TotalCapacityBytes: 0,
		},
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	actions := mock_registry.NewMockSessionActions(mockCtrl)
	sessionRegistry := mock_registry.NewMockSessionRegistry(mockCtrl)
	poolRegistry := mock_registry.NewMockPoolRegistry(mockCtrl)
	allocations := mock_registry.NewMockAllocationRegistry(mockCtrl)
	facade := sessionFacade{
		session: sessionRegistry, actions: actions, pool: poolRegistry, allocations: allocations,
	}

	poolRegistry.EXPECT().GetPool(datamodel.PoolName("pool1")).Return(datamodel.Pool{Name: "pool1"}, nil)
	sessionMutex := mock_store.NewMockMutex(mockCtrl)
	sessionRegistry.EXPECT().GetSessionMutex(initialSession.Name).Return(sessionMutex, nil)
	sessionMutex.EXPECT().Lock(context.TODO())
	sessionRegistry.EXPECT().CreateSession(initialSession).Return(initialSession, nil)
	sessionMutex.EXPECT().Unlock(context.TODO())

	err := facade.CreateSession(initialSession)

	assert.Nil(t, err)
}

func TestSessionFacade_CreateSession_WithBricks_AllocationError(t *testing.T) {
	initialSession := datamodel.Session{
		Name: "foo",
		VolumeRequest: datamodel.VolumeRequest{
			PoolName:           datamodel.PoolName("pool1"),
			TotalCapacityBytes: 2048,
		},
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	actions := mock_registry.NewMockSessionActions(mockCtrl)
	sessionRegistry := mock_registry.NewMockSessionRegistry(mockCtrl)
	poolRegistry := mock_registry.NewMockPoolRegistry(mockCtrl)
	allocations := mock_registry.NewMockAllocationRegistry(mockCtrl)
	facade := sessionFacade{
		session: sessionRegistry, actions: actions, pool: poolRegistry, allocations: allocations,
	}

	poolRegistry.EXPECT().GetPool(datamodel.PoolName("pool1")).Return(datamodel.Pool{Name: "pool1"}, nil)
	sessionMutex := mock_store.NewMockMutex(mockCtrl)
	sessionRegistry.EXPECT().GetSessionMutex(initialSession.Name).Return(sessionMutex, nil)
	sessionMutex.EXPECT().Lock(context.TODO())
	allocationMutex := mock_store.NewMockMutex(mockCtrl)
	allocations.EXPECT().GetAllocationMutex().Return(allocationMutex, nil)
	allocationMutex.EXPECT().Lock(context.TODO())
	allocations.EXPECT().GetPoolInfo(initialSession.VolumeRequest.PoolName).Return(datamodel.PoolInfo{
		Pool: datamodel.Pool{
			Name: "pool1", GranularityBytes: 1024,
		},
	}, nil)
	allocationMutex.EXPECT().Unlock(context.TODO())
	sessionMutex.EXPECT().Unlock(context.TODO())

	err := facade.CreateSession(initialSession)

	assert.Equal(t, "can't allocate for session: foo due to unable to get number of requested bricks (2) for given pool (pool1)", err.Error())
}

func TestSessionFacade_CreateSession_WithBricks_CreateSessionError(t *testing.T) {
	initialSession := datamodel.Session{
		Name: "foo",
		VolumeRequest: datamodel.VolumeRequest{
			PoolName:           datamodel.PoolName("pool1"),
			TotalCapacityBytes: 1024,
		},
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	actions := mock_registry.NewMockSessionActions(mockCtrl)
	sessionRegistry := mock_registry.NewMockSessionRegistry(mockCtrl)
	poolRegistry := mock_registry.NewMockPoolRegistry(mockCtrl)
	allocations := mock_registry.NewMockAllocationRegistry(mockCtrl)
	facade := sessionFacade{
		session: sessionRegistry, actions: actions, pool: poolRegistry, allocations: allocations,
	}

	poolRegistry.EXPECT().GetPool(datamodel.PoolName("pool1")).Return(datamodel.Pool{Name: "pool1"}, nil)
	sessionMutex := mock_store.NewMockMutex(mockCtrl)
	sessionRegistry.EXPECT().GetSessionMutex(initialSession.Name).Return(sessionMutex, nil)
	sessionMutex.EXPECT().Lock(context.TODO())
	allocationMutex := mock_store.NewMockMutex(mockCtrl)
	allocations.EXPECT().GetAllocationMutex().Return(allocationMutex, nil)
	allocationMutex.EXPECT().Lock(context.TODO())
	brickList := []datamodel.Brick{{Device: "sda", BrickHostName: datamodel.BrickHostName("host1")}}
	allocations.EXPECT().GetPoolInfo(initialSession.VolumeRequest.PoolName).Return(datamodel.PoolInfo{
		Pool: datamodel.Pool{
			Name: "pool1", GranularityBytes: 1024,
		},
		AvailableBricks: brickList,
	}, nil)
	allocationList := []datamodel.BrickAllocation{{Brick: brickList[0]}}
	allocations.EXPECT().CreateAllocations(initialSession.Name, brickList).Return(allocationList, nil)
	updatedSession := datamodel.Session{
		Name: "foo",
		VolumeRequest: datamodel.VolumeRequest{
			PoolName:           datamodel.PoolName("pool1"),
			TotalCapacityBytes: 1024,
		},
		ActualSizeBytes:  1024,
		Allocations:      allocationList,
		PrimaryBrickHost: allocationList[0].Brick.BrickHostName,
	}
	returnedSession := datamodel.Session{
		Name:            "foo",
		ActualSizeBytes: 1024,
	}
	sessionRegistry.EXPECT().CreateSession(updatedSession).Return(returnedSession, nil)
	allocationMutex.EXPECT().Unlock(context.TODO())
	fakeErr := errors.New("fake")
	actionChan := make(chan datamodel.SessionAction)
	actions.EXPECT().CreateSessionVolume(context.TODO(), initialSession.Name).Return(actionChan, nil)
	sessionMutex.EXPECT().Unlock(context.TODO())
	go func() {
		actionChan <- datamodel.SessionAction{Error: fakeErr}
		close(actionChan)
	}()

	err := facade.CreateSession(initialSession)

	assert.Equal(t, fakeErr, err)
}

func TestSessionFacade_DeleteSession(t *testing.T) {
	sessionName := datamodel.SessionName("foo")
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	actions := mock_registry.NewMockSessionActions(mockCtrl)
	sessionRegistry := mock_registry.NewMockSessionRegistry(mockCtrl)
	facade := sessionFacade{session: sessionRegistry, actions: actions}
	sessionMutex := mock_store.NewMockMutex(mockCtrl)
	sessionRegistry.EXPECT().GetSessionMutex(sessionName).Return(sessionMutex, nil)
	sessionMutex.EXPECT().Lock(context.TODO())
	initialSession := datamodel.Session{Name: "foo"}
	sessionRegistry.EXPECT().GetSession(sessionName).Return(initialSession, nil)
	updatedSession := datamodel.Session{
		Name: "foo",
		Status: datamodel.SessionStatus{
			DeleteRequested:       true,
			DeleteSkipCopyDataOut: true,
		},
	}
	sessionRegistry.EXPECT().UpdateSession(updatedSession).Return(initialSession, nil)
	actionChan := make(chan datamodel.SessionAction)
	actions.EXPECT().SendSessionAction(context.TODO(), datamodel.SessionDelete, initialSession).Return(actionChan, nil)
	sessionMutex.EXPECT().Unlock(context.TODO())
	fakeErr := errors.New("fake")
	go func() {
		actionChan <- datamodel.SessionAction{Error: fakeErr}
		close(actionChan)
	}()

	err := facade.DeleteSession(sessionName, true)

	assert.Equal(t, fakeErr, err)
}
