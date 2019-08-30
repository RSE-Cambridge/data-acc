package brick_manager_impl

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/config"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacd"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry_impl"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/store"
	"log"
)

func NewBrickManager(keystore store.Keystore) dacd.BrickManager {
	return &brickManager{
		config:               config.GetBrickManagerConfig(config.DefaultEnv),
		brickRegistry:        registry_impl.NewBrickHostRegistry(keystore),
		sessionActions:       registry_impl.NewSessionActionsRegistry(keystore),
		sessionActionHandler: NewSessionActionHandler(keystore),
	}
}

type brickManager struct {
	config               config.BrickManagerConfig
	brickRegistry        registry.BrickHostRegistry
	sessionActions       registry.SessionActions
	sessionActionHandler facade.SessionActionHandler
}

func (bm *brickManager) Hostname() string {
	return string(bm.config.BrickHostName)
}

func (bm *brickManager) Startup() {
	// TODO: should we get the allocation mutex until we are started the keep alive?
	// TODO: add a drain configuration?

	err := bm.brickRegistry.UpdateBrickHost(getBrickHost(bm.config))
	if err != nil {
		log.Panicf("failed to update brick host: %s", err)
	}

	// If we are are enabled, this includes new create session requests
	events, err := bm.sessionActions.GetSessionActionRequests(context.TODO(), bm.config.BrickHostName)

	go func() {
		for event := range events {
			bm.sessionActionHandler.ProcessSessionAction(event)
		}
		log.Println("ERROR: stopped waiting for new Session Actions")
	}()

	// TODO: try to recover all existing filesystems on restart
	//   including a check to make sure all related brick hosts are alive

	// Tell everyone we are listening
	err = bm.brickRegistry.KeepAliveHost(context.TODO(), bm.config.BrickHostName)
	if err != nil {
		log.Panicf("failed to start keep alive host: %s", err)
	}
}

func (bm *brickManager) Shutdown() {
	// Delete the keepalive key, to stop new actions being sent
	// Wait for existing actions by trying to get a lock on all
	// sessions we for which we are the primary brick
	// TODO...
}
