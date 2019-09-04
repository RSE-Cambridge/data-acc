package filesystem

import "github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"

type Provider interface {
	Create(session datamodel.Session) (datamodel.FilesystemStatus, error)
	Restore(session datamodel.Session) error
	Delete(session datamodel.Session) error

	DataCopyIn(session datamodel.Session) error
	DataCopyOut(session datamodel.Session) error

	Mount(session datamodel.Session, attachments datamodel.AttachmentSession) error
	Unmount(session datamodel.Session, attachments datamodel.AttachmentSession) error
}
