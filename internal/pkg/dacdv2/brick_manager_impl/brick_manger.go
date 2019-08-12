package brick_manager_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/dacdv2/brick_manager"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/brick_host"
	"log"
	"os"
)

func NewBrickManager(brickRegistry brick_host.BrickHostRegistry) brick_manager.BrickManager {
	return &brickManager{getHostname(), brickRegistry}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	return hostname
}

type brickManager struct {
	hostname      string
	brickRegistry brick_host.BrickHostRegistry
}

func (bm *brickManager) Hostname() string {
	return bm.hostname
}

func (bm *brickManager) Startup(drainSessions bool) error {
	panic("implement me")
	// * update current brick status
	//   ** error out if removing bricks with existing assignments?
	// * start listening for create sessions (new primary bricks) and session actions
	// * report we are listening with keep-alive
	// * check all brick assignments
	//   ** ensure all brick hosts are up, warn if there are issues
	//   ** if primary brick, refresh ansible run (i.e. recover from host reboot)
}

func (bm *brickManager) Shutdown() error {
	// Delete the keepalive key, to stop new actions being sent
	// Wait for existing actions by trying to get a lock on all
	// sessions we for which we are the primary brick
	panic("implement me")
}
