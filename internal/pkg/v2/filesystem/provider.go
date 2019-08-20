package filesystem

import "github.com/RSE-Cambridge/data-acc/internal/pkg/v2/datamodel"

type Provider interface {
	Create(session datamodel.Session) (datamodel.FilesystemStatus, error)
	Delete(session datamodel.Session) error

	DataCopyIn(session datamodel.Session) error
	DataCopyOut(session datamodel.Session) error

	Mount(session datamodel.Session, attachments datamodel.AttachmentSessionStatus) error
	Unmount(session datamodel.Session, attachments datamodel.AttachmentSessionStatus) error
}
