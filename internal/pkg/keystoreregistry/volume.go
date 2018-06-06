package keystoreregistry

import "github.com/RSE-Cambridge/data-acc/internal/pkg/registry"

func NewVolumeRegistry(keystore Keystore) registry.VolumeRegistry {
	return &VolumeRegistry{keystore}
}

type VolumeRegistry struct {
	keystore Keystore
}

func (*VolumeRegistry) Jobs() ([]registry.Job, error) {
	panic("implement me")
}

func (*VolumeRegistry) Volume(name registry.VolumeName) (registry.Volume, error) {
	panic("implement me")
}

func (*VolumeRegistry) DeleteVolume(volume registry.Volume) {
	panic("implement me")
}

func (*VolumeRegistry) WatchVolumeChanges(volumeName string, callback func(old registry.Volume, new registry.Volume)) error {
	panic("implement me")
}

func (*VolumeRegistry) WaitForVolumeReady(volume registry.Volume) error {
	panic("implement me")
}

func (*VolumeRegistry) WaitForVolumeDataIn(volumeName string) error {
	panic("implement me")
}

func (*VolumeRegistry) WaitForVolumeAttached(volumeName string) error {
	panic("implement me")
}

func (*VolumeRegistry) WaitForVolumeDetached(volumeName string) error {
	panic("implement me")
}

func (*VolumeRegistry) WaitForVolumeDataOut(volumeName string) error {
	panic("implement me")
}

func (*VolumeRegistry) WaitForVolumeDeleted(volumeName string) error {
	panic("implement me")
}
