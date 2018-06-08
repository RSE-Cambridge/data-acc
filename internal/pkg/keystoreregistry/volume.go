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

func (volRegistry *volumeRegistry) AllVolumes() ([]registry.Volume, error) {
	var volumes []registry.Volume
	keyValues, err := volRegistry.keystore.GetAll(volumeKeyPrefix)
	if err != nil {
		return volumes, err
	}
	for _, keyValue := range keyValues {
		var volume registry.Volume
		err = volumeFromKeyValue(keyValue, &volume)
		if err != nil {
			return volumes, nil
		}
		volumes = append(volumes, volume)
	}
	return volumes, nil
}

func (volRegistry *volumeRegistry) Jobs() ([]registry.Job, error) {
	var jobs []registry.Job
	keyValues, err := volRegistry.keystore.GetAll(jobPrefix)
	for _, keyValue := range keyValues {
		var job registry.Job
		err := json.Unmarshal(bytes.NewBufferString(keyValue.Value).Bytes(), &job)
		if err != nil {
			return jobs, err
		}
		jobs = append(jobs, job)
	}
	return jobs, err
}

const jobPrefix = "/job/"

func getJobKey(jobName string) string {
	return fmt.Sprintf("%s%s/", jobPrefix, jobName)
}

func (volRegistry *volumeRegistry) AddJob(job registry.Job) error {
	for _, volumeName := range job.Volumes {
		volume, err := volRegistry.Volume(volumeName)
		if err != nil {
			return err
		}
		// TODO: what other checks are required?
		if volume.State < registry.Registered {
			return fmt.Errorf("must register volume: %s", volume.Name)
		}
	}
	return volRegistry.keystore.Add([]KeyValue{
		{Key: getJobKey(job.Name), Value: toJson(job)},
	})
}

func (volRegistry *volumeRegistry) DeleteJob(jobName string) error {
	keyValue, err := volRegistry.keystore.Get(getJobKey(jobName))
	if err != nil {
		return err
	}
	return volRegistry.keystore.DeleteAll([]KeyValueVersion{keyValue})
}

func (volRegistry *volumeRegistry) UpdateConfiguration(name registry.VolumeName, configurations []registry.Configuration) error {
	updateConfig := func(volume *registry.Volume) error {
		volume.Configurations = configurations
		return nil
	}
	return volRegistry.updateVolume(name, updateConfig)
}

func (volRegistry *volumeRegistry) updateVolume(name registry.VolumeName,
	update func(volume *registry.Volume) error) error {

	keyValue, err := volRegistry.keystore.Get(getVolumeKey(string(name)))
	if err != nil {
		return err
	}

	volume := registry.Volume{}
	err = volumeFromKeyValue(keyValue, &volume)
	if err != nil {
		return nil
	}
	if err := update(&volume); err != nil {
		return err
	}

	keyValue.Value = toJson(volume)
	return volRegistry.keystore.Update([]KeyValueVersion{keyValue})

}
func (volRegistry *volumeRegistry) UpdateState(name registry.VolumeName, state registry.VolumeState) error {
	updateState := func(volume *registry.Volume) error {
		stateDifference := state - volume.State
		if stateDifference != 1 {
			return fmt.Errorf("must update volume to the next state")
		}
		volume.State = state
		return nil
	}
	return volRegistry.updateVolume(name, updateState)
}

const volumeKeyPrefix = "/volume/"

func getVolumeKey(volumeName string) string {
	return fmt.Sprintf("%s%s/", volumeKeyPrefix, volumeName)
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
