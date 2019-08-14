package actions_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/mock_registry"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDacctlActions_CreatePerJobBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_registry.NewMockRegistry(mockCtrl)
	session := mock_registry.NewMockActions(mockCtrl)
	disk := mocks.NewMockDisk(mockCtrl)

	lines := []string{
		`#DW jobdw capacity=4MiB access_mode=striped,private type=scratch`,
		`#DW persistentdw name=myBBname1`,
		`#DW persistentdw name=myBBname2`,
		`#DW swap 4MiB`,
		`#DW stage_in source=/global/cscratch1/filename1 destination=$DW_JOB_STRIPED/filename1 type=file`,
		`#DW stage_in source=/global/cscratch1/filelist type=list`,
		`#DW stage_out source=$DW_JOB_STRIPED/outdir destination=/global/scratch1/outdir type=directory`,
	}
	disk.EXPECT().Lines("jobfile").Return(lines, nil)
	fakeSession := datamodel.Session{Name: "foo"}
	registry.EXPECT().CreateSession(datamodel.Session{
		Name:            "token",
		Owner:           1001,
		Group:           1002,
		CreatedAt:       123,
		MultiJobVolumes: []datamodel.VolumeName{"myBBname1", "myBBname2"},
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
			"DW_PERSISTENT_STRIPED_myBBname1": "/dac/token_persistent_myBBname1",
			"DW_PERSISTENT_STRIPED_myBBname2": "/dac/token_persistent_myBBname2",
		},
	}).Return(fakeSession, nil)
	session.EXPECT().CreateSessionVolume(fakeSession)
	fakeTime = 123

	actions := NewDacctlActions(registry, session, disk)
	err := actions.CreatePerJobBuffer(&mockCliContext{capacity: 2})

	assert.Nil(t, err)
}
