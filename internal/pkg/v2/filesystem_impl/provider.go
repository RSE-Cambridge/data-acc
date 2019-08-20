package filesystem_impl

import (
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/v2/filesystem"
	"log"
)

func NewFileSystemProvider(ansible filesystem.Ansible) filesystem.Provider {
	return &fileSystemProvider{ansible: ansible}
}

type fileSystemProvider struct {
	ansible filesystem.Ansible
}

func (f *fileSystemProvider) Create(session datamodel.Session) (datamodel.FilesystemStatus, error) {
	log.Println("FAKE Create")
	return datamodel.FilesystemStatus{InternalName: "foo", InternalData: "bar"}, nil
}

func (f *fileSystemProvider) Delete(session datamodel.Session) error {
	log.Println("FAKE Delete")
	return nil
}

func (f *fileSystemProvider) DataCopyIn(session datamodel.Session) error {
	log.Println("FAKE DataCopyIn")
	return nil

}

func (f *fileSystemProvider) DataCopyOut(session datamodel.Session) error {
	log.Println("FAKE DataCopyOut")
	return nil

}

func (f *fileSystemProvider) Mount(session datamodel.Session, attachments datamodel.AttachmentSession) datamodel.AttachmentSessionStatus {
	log.Println("FAKE Mount")
	return datamodel.AttachmentSessionStatus{}

}

func (f *fileSystemProvider) Unmount(session datamodel.Session, attachments datamodel.AttachmentSession) datamodel.AttachmentSessionStatus {
	log.Println("FAKE Unmount")
	return datamodel.AttachmentSessionStatus{}
}
