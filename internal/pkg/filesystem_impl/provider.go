package filesystem_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/filesystem"
	"math/rand"
	"time"
)

func NewFileSystemProvider(ansible filesystem.Ansible) filesystem.Provider {
	return &fileSystemProvider{ansible: ansible}
}

type fileSystemProvider struct {
	ansible filesystem.Ansible
	// TODO: proper config object
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var source = rand.NewSource(time.Now().UnixNano())
var randGenerator = rand.New(source)

func GetNewUUID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[randGenerator.Int63()%int64(len(letters))]
	}
	return string(b)
}

func (f *fileSystemProvider) Create(session datamodel.Session) (datamodel.FilesystemStatus, error) {
	session.FilesystemStatus = datamodel.FilesystemStatus{
		InternalName: GetNewUUID(),
		InternalData: "",
	}
	err := executeAnsibleSetup(session.FilesystemStatus.InternalName, session.AllocatedBricks, true)
	return session.FilesystemStatus, err
}

func (f *fileSystemProvider) Restore(session datamodel.Session) error {
	return executeAnsibleSetup(session.FilesystemStatus.InternalName, session.AllocatedBricks, false)
}

func (f *fileSystemProvider) Delete(session datamodel.Session) error {
	return executeAnsibleTeardown(session.FilesystemStatus.InternalName, session.AllocatedBricks)
}

func (f *fileSystemProvider) DataCopyIn(session datamodel.Session) error {
	for _, dataCopy := range session.StageInRequests {
		err := processDataCopy(session, dataCopy)
		if err != nil {
			return err
		}
	}
	return nil

}

func (f *fileSystemProvider) DataCopyOut(session datamodel.Session) error {
	for _, dataCopy := range session.StageOutRequests {
		err := processDataCopy(session, dataCopy)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *fileSystemProvider) Mount(session datamodel.Session, attachments datamodel.AttachmentSession) error {
	return mount(Lustre, session.Name, session.VolumeRequest.MultiJob, session.FilesystemStatus.InternalName,
		session.PrimaryBrickHost, attachments, session.Owner, session.Group)

}

func (f *fileSystemProvider) Unmount(session datamodel.Session, attachments datamodel.AttachmentSession) error {
	return unmount(Lustre, session.Name, session.VolumeRequest.MultiJob, session.FilesystemStatus.InternalName,
		session.PrimaryBrickHost, attachments)
}
