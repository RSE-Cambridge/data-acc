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
	return fmt.Sprintf("/volume/%s/", volumeName)
}

func (volRegistry *volumeRegistry) AddVolume(volume registry.Volume) error {
	return volRegistry.keystore.Add([]KeyValue{{
		Key:   getVolumeKey(string(volume.Name)),
		Value: toJson(volume),
	}})
}

func volumeFromKeyValue(keyValue KeyValueVersion, volume *registry.Volume) error {
	return json.Unmarshal(bytes.NewBufferString(keyValue.Value).Bytes(), &volume)
}

func (volRegistry *volumeRegistry) Volume(name registry.VolumeName) (registry.Volume, error) {
	volume := registry.Volume{}
	keyValue, err := volRegistry.keystore.Get(getVolumeKey(string(name)))
	if err != nil {
		return volume, err
	}
	err = volumeFromKeyValue(keyValue, &volume)
	if err != nil {
		return volume, nil
	}
	return volume, nil
}

func (volRegistry *volumeRegistry) DeleteVolume(name registry.VolumeName) error {
	keyValue, err := volRegistry.keystore.Get(getVolumeKey(string(name)))
	if err != nil {
		return err
	}
	return volRegistry.keystore.DeleteAll([]KeyValueVersion{keyValue})
}

func (volRegistry *volumeRegistry) WatchVolumeChanges(volumeName string,
	callback func(old *registry.Volume, new *registry.Volume)) error {
	key := getVolumeKey(volumeName)
	volRegistry.keystore.WatchPrefix(key, func(old *KeyValueVersion, new *KeyValueVersion) {
		// TODO watching prefix but really just want to watch a key
		oldVolume := &registry.Volume{}
		newVolume := &registry.Volume{}
		if old != nil {
			volumeFromKeyValue(*old, oldVolume)
		}
		if new != nil {
			volumeFromKeyValue(*new, newVolume)
		}
		callback(oldVolume, newVolume)
	})
	return nil // TODO check key is present
}

func (volRegistry *volumeRegistry) Jobs() ([]registry.Job, error) {
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
