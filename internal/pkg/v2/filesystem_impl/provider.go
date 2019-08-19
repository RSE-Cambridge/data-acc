package filesystem_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/filesystem"
)

func NewFileSystemProvider(ansible filesystem.Ansible) filesystem.Provider {
	return &fileSystemProvider{ansible:ansible}
}

type fileSystemProvider struct {
	ansible filesystem.Ansible
}

func (f *fileSystemProvider) Create(session datamodel.Session) (datamodel.FilesystemStatus, error) {
	panic("implement me")
}

func (f *fileSystemProvider) Delete(session datamodel.Session) error {
	panic("implement me")
}

func (f *fileSystemProvider) DataCopyIn(session datamodel.Session) error {
	panic("implement me")
}

func (f *fileSystemProvider) DataCopyOut(session datamodel.Session) error {
	panic("implement me")
}

func (f *fileSystemProvider) Mount(session datamodel.Session, attachments datamodel.AttachmentSession) datamodel.AttachmentSessionStatus {
	panic("implement me")
}

func (f *fileSystemProvider) Unmount(session datamodel.Session, attachments datamodel.AttachmentSession) datamodel.AttachmentSessionStatus {
	panic("implement me")
}
