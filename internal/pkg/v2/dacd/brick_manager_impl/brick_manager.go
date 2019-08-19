package brick_manager_impl

import (
	"context"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacd"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/dacd/config"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/facade"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/registry"
	"log"
)

func NewBrickManager(brickRegistry registry.BrickHostRegistry, handler facade.SessionActionHandler) dacd.BrickManager {
	return &brickManager{
		config:               config.GetBrickManagerConfig(config.DefaultEnv),
		brickRegistry:        brickRegistry,
		sessionActionHandler: handler,
	}
}

type brickManager struct {
	config               config.BrickManagerConfig
	brickRegistry        registry.BrickHostRegistry
	sessionActionHandler facade.SessionActionHandler
}

func (bm *brickManager) Hostname() string {
	return string(bm.config.BrickHostName)
}

func (bm *brickManager) Startup(drainSessions bool) error {
	err := bm.brickRegistry.UpdateBrickHost(getBrickHost(bm.config))
	if err != nil {
		return err
	}

	// If we are are enabled, this includes new create session requests
	events, err := bm.brickRegistry.GetSessionActions(context.TODO(), bm.config.BrickHostName)

	go func() {
		for event := range events {
			bm.sessionActionHandler.ProcessSessionAction(event)
		}
		log.Println("ERROR: stopped waiting for new Session Actions")
	}()

	// TODO: try to recover all existing filesystems on restart
	//   including a check to make sure all related brick hosts are alive

	// Tell everyone we are listening
	return bm.brickRegistry.KeepAliveHost(context.TODO(), bm.config.BrickHostName)
}

func (bm *brickManager) Shutdown() error {
	// Delete the keepalive key, to stop new actions being sent
	// Wait for existing actions by trying to get a lock on all
	// sessions we for which we are the primary brick
	panic("implement me")
}
