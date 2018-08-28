package keystoreregistry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"log"
	"math/rand"
	"sync"
	"time"
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

func (volRegistry *volumeRegistry) Job(jobName string) (registry.Job, error) {
	var job registry.Job // TODO return a pointer instead?
	keyValue, err := volRegistry.keystore.Get(getJobKey(jobName))
	if err != nil {
		return job, err
	}
	err = json.Unmarshal(bytes.NewBufferString(keyValue.Value).Bytes(), &job)
	if err != nil {
		return job, err
	}
	return job, nil
}

func (volRegistry *volumeRegistry) AddJob(job registry.Job) error {
	for _, volumeName := range job.MultiJobVolumes {
		volume, err := volRegistry.Volume(volumeName)
		if err != nil {
			return err
		}
		// TODO: what other checks are required?
		if volume.State < registry.Registered {
			return fmt.Errorf("must register volume: %s", volume.Name)
		}
	}
	if job.JobVolume != "" {
		volume, err := volRegistry.Volume(job.JobVolume)
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

func (volRegistry *volumeRegistry) JobAttachHosts(jobName string, hosts []string) error {
	keyValue, err := volRegistry.keystore.Get(getJobKey(jobName))
	if err != nil {
		return err
	}
	var job registry.Job
	err = json.Unmarshal(bytes.NewBufferString(keyValue.Value).Bytes(), &job)
	if err != nil {
		return err
	}

	// TODO validate hostnames?
	job.AttachHosts = hosts
	keyValue.Value = toJson(job)

	return volRegistry.keystore.Update([]KeyValueVersion{keyValue})
}

func (volRegistry *volumeRegistry) UpdateVolumeAttachments(name registry.VolumeName,
	attachments map[string]registry.Attachment) error {

	update := func(volume *registry.Volume) error {
		if volume.Attachments == nil {
			volume.Attachments = attachments
		} else {
			for key, value := range attachments {
				volume.Attachments[key] = value
			}
		}
		return nil
	}
	return volRegistry.updateVolume(name, update)
}

func (volRegistry *volumeRegistry) DeleteVolumeAttachments(name registry.VolumeName, hostnames []string) error {

	update := func(volume *registry.Volume) error {
		if volume.Attachments == nil {
			return errors.New("no attachments to delete")
		} else {
			for _, hostname := range hostnames {
				_, ok := volume.Attachments[hostname]
				if ok {
					delete(volume.Attachments, hostname)
				} else {
					return fmt.Errorf("unable to find attachment for volume %s and host %s", name, hostname)
				}
			}
		}
		return nil
	}
	return volRegistry.updateVolume(name, update)
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
		if stateDifference != 1 && state != registry.Error && state != registry.DeleteRequested {
			return fmt.Errorf("must update volume %s to the next state, current state: %s",
				volume.Name, volume.State)
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GetNewUUID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Int63()%int64(len(letters))]
	}
	return string(b)
}

func (volRegistry *volumeRegistry) AddVolume(volume registry.Volume) error {
	// TODO: both uuid and client port might clash, need to check they don't!!
	volume.UUID = GetNewUUID()
	volume.ClientPort = rand.Intn(50000) + 10000
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
	callback func(old *registry.Volume, new *registry.Volume) bool) error {
	key := getVolumeKey(volumeName)
	ctxt, cancelFunc := context.WithCancel(context.Background())
	volRegistry.keystore.WatchKey(ctxt, key, func(old *KeyValueVersion, new *KeyValueVersion) {
		oldVolume := &registry.Volume{}
		newVolume := &registry.Volume{}
		if old != nil {
			volumeFromKeyValue(*old, oldVolume)
		}
		if new != nil {
			volumeFromKeyValue(*new, newVolume)
		}
		if callback(oldVolume, newVolume) {
			log.Println("stopping watching volume", volumeName)
			cancelFunc()
		}
	})
	return nil // TODO check key is present
}

func (volRegistry *volumeRegistry) WaitForState(volumeName registry.VolumeName, state registry.VolumeState) error {
	log.Println("Start waiting for volume", volumeName, "to reach state", state)
	err := volRegistry.WaitForCondition(volumeName, func(old *registry.Volume, new *registry.Volume) bool {
		return new.State == state
	})
	log.Println("Stopped waiting for volume", volumeName, "to reach state", state, err)
	return err
}

func (volRegistry *volumeRegistry) WaitForCondition(volumeName registry.VolumeName,
	condition func(old *registry.Volume, new *registry.Volume) bool) error {

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	ctxt, cancelFunc := context.WithTimeout(context.Background(), time.Minute*10)
	// TODO should we always need to call cancel? or is timeout enough?

	err := fmt.Errorf("error waiting for volume %s to meet supplied condition", volumeName)

	var finished bool
	volRegistry.keystore.WatchKey(ctxt, getVolumeKey(string(volumeName)),
		func(old *KeyValueVersion, new *KeyValueVersion) {
			if old == nil && new == nil {
				// TODO: attempt to signal error on timeout, should move to channel!!
				cancelFunc()
				if !finished {
					// at the end we always get called with nil, nil
					// but sometimes we will have already found the condition
					waitGroup.Done()
				}
			}
			oldVolume := &registry.Volume{}
			newVolume := &registry.Volume{}
			if old != nil {
				volumeFromKeyValue(*old, oldVolume)
			}
			if new != nil {
				volumeFromKeyValue(*new, newVolume)
			}

			if condition(oldVolume, newVolume) {
				err = nil
				cancelFunc()
				waitGroup.Done()
				finished = true
			}
		})

	// check we have not already hit the condition
	volume, err := volRegistry.Volume(volumeName)
	if err != nil {
		// NOTE this forces the volume to existing before you wait, seems OK
		return err
	}
	log.Printf("About to wait for condition on volume: %s", volume)
	if condition(&volume, &volume) {
		cancelFunc()
		return nil
	}

	// TODO do we get stuck in a forever loop here when we hit the timeout above?
	waitGroup.Wait()
	return err
}
