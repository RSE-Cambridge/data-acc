package fakewarp

import (
	"log"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestParseJobRequest(t *testing.T) {
	jobRequest := []string{
		`#BB create_persistent name=myBBname capacity=100GB access_mode=striped type=scratch`,
		`#BB create_persistent name=myBBname capacity=100GB access_mode=ldbalance type=cache`,
		`#BB destroy_persistent name=myBBname`,
		`#DW persistentdw name=myBBname1`,
		`#DW persistentdw name=myBBname2`,
		`#DW persistentdw name=myBBname2`,
		`#DW jobdw capacity=10GB access_mode=striped type=scratch`,
		`#DW jobdw capacity=2TB access_mode=private type=scratch`,
		`#DW jobdw capacity=4TiB access_mode=striped,private type=scratch`,
		`#DW jobdw capacity=42GiB access_mode=ldbalance type=cache pfs=/global/scratch1/john`,
		`#DW swap 3TiB`,
		`#DW stage_in source=/global/cscratch1/filename1 destination=$DW_JOB_STRIPED/filename1 type=file`,
		`#DW stage_out source=$DW_JOB_STRIPED/outdir destination=/global/scratch1/outdir type=directory`,
	}
	if cmds, err := parseJobRequest(jobRequest); err != nil {
		log.Fatal(err)
	} else {
		assert.Equal(t, 13, len(jobRequest)) // TODO should check returned values!!
		for _, cmd := range cmds {
			log.Printf("Cmd: %T Args: %s\n", cmd, cmd)
		}
	}
}
