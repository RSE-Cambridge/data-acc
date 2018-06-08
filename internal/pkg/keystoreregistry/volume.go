package keystoreregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
)

func NewVolumeRegistry(keystore Keystore) registry.VolumeRegistry {
	return &volumeRegistry{keystore}
}

type volumeRegistry struct {
	keystore Keystore
}

func getVolumeKey(volumeName string) string {
	return fmt.Sprintf("/volume/%s", volumeName)
}

func (volRegistry *volumeRegistry) AddVolume(volume registry.Volume) error {
	return volRegistry.keystore.Add([]KeyValue{{
		Key:   getVolumeKey(string(volume.Name)),
		Value: toJson(volume),
	}})
}

func (volRegistry *volumeRegistry) Volume(name registry.VolumeName) (registry.Volume, error) {
	volume := registry.Volume{}
	keyValue, err := volRegistry.keystore.Get(getVolumeKey(string(name)))
	if err != nil {
		return volume, err
	}
	json.Unmarshal(bytes.NewBufferString(keyValue.Value).Bytes(), &volume)
	return volume, nil
}

func (volRegistry *volumeRegistry) DeleteVolume(name registry.VolumeName) error {
	keyValue, err := volRegistry.keystore.Get(getVolumeKey(string(name)))
	if err != nil {
		return err
	}
	return volRegistry.keystore.DeleteAll([]KeyValueVersion{keyValue})
}

func (volRegistry *volumeRegistry) Jobs() ([]registry.Job, error) {
	panic("implement me")
}

func (volRegistry *volumeRegistry) WatchVolumeChanges(volumeName string, callback func(old registry.Volume, new registry.Volume)) error {
	panic("implement me")
}

func (volRegistry *volumeRegistry) WaitForVolumeReady(volume registry.Volume) error {
	panic("implement me")
}

func (volRegistry *volumeRegistry) WaitForVolumeDataIn(volumeName string) error {
	panic("implement me")
}

func (volRegistry *volumeRegistry) WaitForVolumeAttached(volumeName string) error {
	panic("implement me")
}

func (volRegistry *volumeRegistry) WaitForVolumeDetached(volumeName string) error {
	panic("implement me")
}

func (volRegistry *volumeRegistry) WaitForVolumeDataOut(volumeName string) error {
	panic("implement me")
}

func (volRegistry *volumeRegistry) WaitForVolumeDeleted(volumeName string) error {
	panic("implement me")
}
