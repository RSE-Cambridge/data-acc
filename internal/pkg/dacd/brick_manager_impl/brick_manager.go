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
		sessionRegistry:      registry_impl.NewSessionRegistry(keystore),
		sessionActions:       registry_impl.NewSessionActionsRegistry(keystore),
		sessionActionHandler: NewSessionActionHandler(keystore),
	}
}

type brickManager struct {
	config               config.BrickManagerConfig
	brickRegistry        registry.BrickHostRegistry
	sessionRegistry      registry.SessionRegistry
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

	// Assume we got restarted, first try to finish all pending actions
	bm.completePendingActions()

	// If we were restarted, likely no one is listening for pending actions any more
	// so don'y worry the above pending actions may have failed due to not restoring sessions first
	bm.restoreSessions()

	// Tell everyone we are listening
	err = bm.brickRegistry.KeepAliveHost(context.TODO(), bm.config.BrickHostName)
	if err != nil {
		log.Panicf("failed to start keep alive host: %s", err)
	}

	// Process any events, given others know we are alive
	go func() {
		for event := range events {
			// TODO: we could limit the number of workers
			go bm.sessionActionHandler.ProcessSessionAction(event)
		}
		log.Println("ERROR: stopped waiting for new Session Actions")
	}()
}

func (bm *brickManager) completePendingActions() {
	// Assume the service has been restarted, lets
	// retry any actions that haven't been completed
	// making the assumption that actions are idempotent
	actions, err := bm.sessionActions.GetOutstandingSessionActionRequests(bm.config.BrickHostName)
	if err != nil {
		log.Fatalf("unable to get outstanding session action requests due to: %s", err.Error())
	}

	for _, action := range actions {
		// TODO: what about the extra response if no one is listening any more?
		bm.sessionActionHandler.ProcessSessionAction(action)
	}
}

func (bm *brickManager) restoreSessions() {
	// In case the server was restarted, double check everything is up
	// If marked deleted, and not already deleted, delete it
	sessions, err := bm.sessionRegistry.GetAllSessions()
	if err != nil {
		log.Panicf("unable to fetch all sessions due to: %s", err)
	}
	for _, session := range sessions {
		hasLocalBrick := false
		for _, brick := range session.AllocatedBricks {
			if brick.BrickHostName == bm.config.BrickHostName {
				hasLocalBrick = true
			}
		}
		if !hasLocalBrick {
			continue
		}

		if session.Status.FileSystemCreated && !session.Status.DeleteRequested {
			// If we have previously finished creating the session,
			// and we don't have a pending delete, try to restore the session
			log.Println("Restoring session with local brick", session.Name)
			go bm.sessionActionHandler.RestoreSession(session)
		} else {
			// TODO: should we just do the delete here?
			log.Printf("WARNING session in strange state: %+v\n", session)
		}
	}
}

func (bm *brickManager) Shutdown() {
	// Delete the keepalive key, to stop new actions being sent
	// Wait for existing actions by trying to get a lock on all
	// sessions we for which we are the primary brick
	// TODO...
}
