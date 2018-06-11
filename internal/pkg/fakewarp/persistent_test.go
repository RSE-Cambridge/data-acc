package fakewarp

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type mockCliContext struct{}

func (c *mockCliContext) String(name string) string {
	switch name {
	case "capacity":
		return "pool1:42GiB"
	default:
		return ""
	}
}

func (c *mockCliContext) Int(name string) int {
	return 42 + len(name)
}

func TestCreatePersistentBufferReturnsError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockObj := mocks.NewMockVolumeRegistry(mockCtrl)
	mockObj.EXPECT().AddVolume(gomock.Any()) // TODO
	mockObj.EXPECT().AddJob(gomock.Any())
	mockCtxt := mockCliContext{}

	if actual, err := CreatePersistentBuffer(&mockCtxt, mockObj); err != nil {
		assert.EqualValues(t, "unable to create buffer", fmt.Sprint(err))
	} else {
		assert.EqualValues(t, "", actual) // TODO
	}
}

func TestParseJobRequest(t *testing.T) {
	jobRequest := []string{
		`#BB create_persistent name=myBBname capacity=100GB access_mode=striped type=scratch`,
		`#BB create_persistent name=myBBname capacity=100GB access_mode=ldbalance type=cache`,
		`#BB destroy_persistent name=myBBname`,
		`#DW persistentdw name=myBBname1`,
		`#DW persistentdw name=myBBname2`,
		`#DW persistentdw name=myBBname2`,
		`#DW jobdw capacity=10GB access_mode=striped type=scratch`,
		`#DW jobdw capacity=10GB access_mode=private type=scratch`,
		`#DW jobdw capacity=4TiB access_mode=striped,private type=scratch`,
		`#DW jobdw capacity=42GiB access_mode=ldbalance type=cache pfs=/global/scratch1/john`,
		`#DW swap 3TiB`,
		`#DW stage_in source=/global/cscratch1/filename1 destination=$DW_JOB_STRIPED/filename1 type=file`,
		`#DW stage_out source=$DW_JOB_STRIPED/outdir destination=/global/scratch1/outdir type=directory`,
	}
	if err := parseJobRequest(jobRequest); err != nil {
		log.Fatal(err)
	}

}
