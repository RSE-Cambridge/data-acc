package actions_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mock_fileio"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_facade"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDacctlActions_ValidateJob_BadInput(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)
	disk := mock_fileio.NewMockDisk(mockCtrl)

	lines := []string{`#DW bad cmd`}
	disk.EXPECT().Lines("jobfile").Return(lines, nil)
	actions := dacctlActions{session: session, disk: disk}
	err := actions.ValidateJob(&mockCliContext{
		strings: map[string]string{
			"job": "jobfile",
		},
	})

	assert.Equal(t, "unrecognised command: bad with arguments: [cmd]", err.Error())

	err = actions.ValidateJob(&mockCliContext{})
	assert.Equal(t, "Please provide these required parameters: job", err.Error())
}

func TestDacctlActions_CreatePerJobBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	session := mock_facade.NewMockSession(mockCtrl)
	disk := mock_fileio.NewMockDisk(mockCtrl)

	lines := []string{
		`#DW jobdw capacity=4MiB access_mode=striped,private type=scratch`,
		`#DW persistentdw name=myBBname2`,
		`#DW persistentdw name=myBBname1`,
		`#DW swap 4MiB`,
		`#DW stage_in source=/global/cscratch1/filename1 destination=$DW_JOB_STRIPED/filename1 type=file`,
		`#DW stage_in source=/global/cscratch1/filelist type=list`,
		`#DW stage_out source=$DW_JOB_STRIPED/outdir destination=/global/scratch1/outdir type=directory`,
	}
	disk.EXPECT().Lines("jobfile").Return(lines, nil)
	session.EXPECT().CreateSession(datamodel.Session{
		Name:                "token",
		Owner:               1001,
		Group:               1002,
		CreatedAt:           123,
		MultiJobAttachments: []datamodel.SessionName{"myBBname1", "myBBname2"},
		StageInRequests: []datamodel.DataCopyRequest{
			{
				SourceType:  datamodel.File,
				Source:      "/global/cscratch1/filename1",
				Destination: "$DW_JOB_STRIPED/filename1",
			},
			{
				SourceType: datamodel.List,
				Source:     "/global/cscratch1/filelist",
			},
		},
		StageOutRequests: []datamodel.DataCopyRequest{
			{
				SourceType:  datamodel.Directory,
				Source:      "$DW_JOB_STRIPED/outdir",
				Destination: "/global/scratch1/outdir",
			},
		},
		VolumeRequest: datamodel.VolumeRequest{
			Caller:             "caller",
			PoolName:           "pool1",
			TotalCapacityBytes: 2147483648,
			Access:             datamodel.PrivateAndStriped,
			Type:               datamodel.Scratch,
			SwapBytes:          4194304,
		},
		Paths: map[string]string{
			"DW_JOB_PRIVATE":                  "/dac/token_job_private",
			"DW_JOB_STRIPED":                  "/dac/token_job/global",
			"DW_PERSISTENT_STRIPED_myBBname1": "/dac/token_persistent_myBBname1/global",
			"DW_PERSISTENT_STRIPED_myBBname2": "/dac/token_persistent_myBBname2/global",
		},
	}).Return(nil)

	fakeTime = 123
	actions := dacctlActions{session: session, disk: disk}
	err := actions.CreatePerJobBuffer(getMockCliContext(2))

	assert.Nil(t, err)
}
