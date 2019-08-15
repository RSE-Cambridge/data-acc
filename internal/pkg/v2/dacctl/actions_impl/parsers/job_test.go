package parsers

import (
	"errors"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestParseJobFile(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileIO := mocks.NewMockDisk(mockCtrl)
	mockFileIO.EXPECT().Lines("testfile").Return(nil, errors.New("asdf"))

	lines, err := ParseJobFile(mockFileIO, "testfile")
	assert.Equal(t, jobSummary{}, lines)
	assert.Equal(t, "asdf", err.Error())

	mockFileIO = mocks.NewMockDisk(mockCtrl)
	mockFileIO.EXPECT().Lines("testfile").Return([]string{`#DW swap asdf`}, nil)

	lines, err = ParseJobFile(mockFileIO, "testfile")
	assert.Equal(t, jobSummary{}, lines)
	assert.Equal(t, "unable to parse size: asdf", err.Error())

}

func TestParseJobRequest(t *testing.T) {
	jobRequest := []string{
		`#BB create_persistent name=myBBname capacity=100GB access_mode=striped type=scratch`,
		`#BB create_persistent name=myBBname capacity=1073741824 access_mode=striped type=cache`,
		`#BB destroy_persistent name=myBBname`,
		`#DW persistentdw name=myBBname1`,
		`#DW persistentdw name=myBBname2`,
		`#DW persistentdw name=myBBname2`,
		`#DW jobdw capacity=10GB access_mode=striped type=scratch`,
		`#DW jobdw capacity=2TB access_mode=private type=scratch`,
		`#DW jobdw capacity=4TiB access_mode=striped,private type=scratch`,
		`#BB jobdw capacity=42GiB access_mode=ldbalance type=cache pfs=/global/scratch1/john`,
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

func TestGetJobSummary(t *testing.T) {
	lines := []string{
		`#DW persistentdw name=myBBname1`,
		`#DW persistentdw name=myBBname2`,
		`#DW jobdw capacity=4MiB access_mode=striped,private type=scratch`,
		`#DW swap 4MB`,
		`#DW stage_in source=/global/cscratch1/filename1 destination=$DW_JOB_STRIPED/filename1 type=file`,
		`#DW stage_in source=/global/cscratch1/filename2 destination=$DW_JOB_STRIPED/filename2 type=file`,
		`#DW stage_out source=$DW_JOB_STRIPED/outdir destination=/global/scratch1/outdir type=directory`,
		`skipping other lines that we`,
		`don't understand`,
	}
	result, err := getJobSummary(lines)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(result.DataIn))
	assert.Equal(t, 1, len(result.DataOut))
	assert.EqualValues(t, "/global/cscratch1/filename1", result.DataIn[0].Source)
	assert.EqualValues(t, "/global/cscratch1/filename2", result.DataIn[1].Source)
	assert.EqualValues(t, "$DW_JOB_STRIPED/outdir", result.DataOut[0].Source)

	assert.Equal(t, 2, len(result.Attachments))
	assert.Equal(t, datamodel.SessionName("myBBname1"), result.Attachments[0])
	assert.Equal(t, datamodel.SessionName("myBBname2"), result.Attachments[1])

	assert.Equal(t, 4194304, result.PerJobBuffer.CapacityBytes)
	assert.Equal(t, 4000000, result.Swap.SizeBytes)
}

func TestGetJobSummary_Errors(t *testing.T) {
	lines := []string{`#DW bad_command asdf=asdf`}
	result, err := getJobSummary(lines)
	assert.Equal(t, "unrecognised command: bad_command with arguments: [asdf=asdf]", err.Error())
	assert.Nil(t, result.PerJobBuffer)

	lines = []string{`#DW swap 1B asdf`}
	result, err = getJobSummary(lines)
	assert.Equal(t, "unable to parse swap command: #DW swap 1B asdf", err.Error())
	assert.Nil(t, result.PerJobBuffer)

	lines = []string{`#DW swap 1B`}
	result, err = getJobSummary(lines)
	assert.Equal(t, "unable to parse size: 1B", err.Error())
	assert.Nil(t, result.PerJobBuffer)

	lines = []string{`#DW jobdw capacity=4B access_mode=striped,private type=scratch`}
	result, err = getJobSummary(lines)
	assert.Equal(t, "unable to parse size: 4B", err.Error())
	assert.Nil(t, result.PerJobBuffer)

	lines = []string{`#BB create_persistent name=myBBname capacity=100B access_mode=striped type=scratch`}
	result, err = getJobSummary(lines)
	assert.Equal(t, "unable to parse size: 100B", err.Error())
	assert.Nil(t, result.PerJobBuffer)

	lines = []string{`#DW swap 1MB`, `#DW swap 2MB`}
	result, err = getJobSummary(lines)
	assert.Equal(t, "only one swap request allowed", err.Error())
	assert.Nil(t, result.PerJobBuffer)

	lines = []string{
		`#DW jobdw capacity=4MiB access_mode=private type=scratch`,
		`#DW jobdw capacity=5MiB access_mode=striped type=scratch`,
	}
	result, err = getJobSummary(lines)
	assert.Equal(t, "only one per job buffer allowed", err.Error())
	assert.Nil(t, result.PerJobBuffer)
}
