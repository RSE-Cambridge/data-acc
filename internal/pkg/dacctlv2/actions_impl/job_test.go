package actions_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/mock_session"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/data/model"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDacctlActions_CreatePerJobBuffer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	registry := mock_session.NewMockRegistry(mockCtrl)
	session := mock_session.NewMockActions(mockCtrl)
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
	fakeSession := model.Session{Name: "foo"}
	registry.EXPECT().CreateSession(model.Session{
		Name:            "token",
		Owner:           1001,
		Group:           1002,
		CreatedAt:       123,
		MultiJobVolumes: []model.VolumeName{"myBBname1", "myBBname2"},
		StageInRequests: []model.DataCopyRequest{
			{
				SourceType:  model.File,
				Source:      "/global/cscratch1/filename1",
				Destination: "$DW_JOB_STRIPED/filename1",
			},
			{
				SourceType: model.List,
				Source:     "/global/cscratch1/filelist",
			},
		},
		StageOutRequests: []model.DataCopyRequest{
			{
				SourceType:  model.Directory,
				Source:      "$DW_JOB_STRIPED/outdir",
				Destination: "/global/scratch1/outdir",
			},
		},
		VolumeRequest: model.VolumeRequest{
			Caller:             "caller",
			PoolName:           "pool1",
			TotalCapacityBytes: 2147483648,
			Access:             model.PrivateAndStriped,
			Type:               model.Scratch,
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
