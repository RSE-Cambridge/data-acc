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

func findAttachment(attachments []registry.Attachment,
	hostname string, jobName string) (*registry.Attachment, bool) {
	for _, candidate := range attachments {
		if candidate.Hostname == hostname && candidate.Job == jobName {
			// TODO: double check for duplicate match?
			return &candidate, true
		}
	}
	return nil, false
}

func mergeAttachments(oldAttachments []registry.Attachment, updates []registry.Attachment) []registry.Attachment {
	var newAttachments []registry.Attachment
	for _, update := range updates {
		newAttachments = append(newAttachments, update)
	}

	// add any existing attachments that don't match an update
	for _, oldAttachment := range oldAttachments {
		_, ok := findAttachment(
			updates, oldAttachment.Hostname, oldAttachment.Job)
		if !ok {
			newAttachments = append(newAttachments, oldAttachment)
		}
	}
	return newAttachments
}

func (volRegistry *volumeRegistry) UpdateVolumeAttachments(name registry.VolumeName,
	updates []registry.Attachment) error {
	update := func(volume *registry.Volume) error {
		volume.Attachments = mergeAttachments(volume.Attachments, updates)
		return nil
	}
	return volRegistry.updateVolume(name, update)
}

func (volRegistry *volumeRegistry) DeleteVolumeAttachments(name registry.VolumeName, hostnames []string, jobName string) error {

	update := func(volume *registry.Volume) error {
		if volume.Attachments == nil {
			return errors.New("no attachments to delete")
		} else {
			numberRemoved := removeAttachments(volume, jobName, hostnames)
			if numberRemoved != len(hostnames) {
				return fmt.Errorf("unable to find all attachments for volume %s", name)
			}
		}
		return nil
	}
	return volRegistry.updateVolume(name, update)
}

func removeAttachments(volume *registry.Volume, jobName string, hostnames []string) int {
	var newAttachments []registry.Attachment
	for _, attachment := range volume.Attachments {
		remove := false
		if attachment.Job == jobName {
			for _, host := range hostnames {
				if attachment.Hostname == host {
					remove = true
					break
				}
			}
		}
		if !remove {
			newAttachments = append(newAttachments, attachment)
		}
	}
	numberRemoved := len(volume.Attachments) - len(newAttachments)
	volume.Attachments = newAttachments
	return numberRemoved
}

func (volRegistry *volumeRegistry) updateVolume(name registry.VolumeName,
	update func(volume *registry.Volume) error) error {

	// TODO: if we restructure attachments into separate keys, we can probably ditch this mutex
	mutex, err := volRegistry.keystore.NewMutex(getVolumeKey(string(name)))
	if err != nil {
		return err
	}
	if err := mutex.Lock(context.TODO()); err != nil {
		return err
	}
	defer mutex.Unlock(context.TODO())

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

func (volRegistry *volumeRegistry) VolumeOperationMutex(name registry.VolumeName) (registry.Mutex, error) {
	return volRegistry.keystore.NewMutex(fmt.Sprintf("operation_%s", name))
}

func (volRegistry *volumeRegistry) UpdateState(name registry.VolumeName, state registry.VolumeState) error {
	updateState := func(volume *registry.Volume) error {
		stateDifference := state - volume.State
		if stateDifference != 1 && state != registry.Error && state != registry.DeleteRequested {
			return fmt.Errorf("must update volume %s to the next state, current state: %s",
				volume.Name, volume.State)
		}
		volume.State = state
		if state == registry.BricksAllocated {
			// From this point onwards, we know bricks might need to be cleaned up
			volume.HadBricksAssigned = true
		}
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

func (volRegistry *volumeRegistry) GetVolumeChanges(ctx context.Context, volume registry.Volume) registry.VolumeChangeChan {
	// TODO: we should watch from the version of the passed in volume
	key := getVolumeKey(string(volume.Name))
	rawEvents := volRegistry.keystore.Watch(ctx, key, false)

	events := make(chan registry.VolumeChange)

	go func() {
		defer close(events)
		if rawEvents == nil {
			return
		}
		for rawEvent := range rawEvents {
			if rawEvent.Err != nil {
				events <- registry.VolumeChange{Err: rawEvent.Err}
				continue
			}

			event := registry.VolumeChange{
				IsDelete: rawEvent.IsDelete,
				Old:      nil,
				New:      nil,
			}
			if rawEvent.Old != nil {
				oldVolume := &registry.Volume{}
				if err := volumeFromKeyValue(*rawEvent.Old, oldVolume); err != nil {
					event.Err = err
				} else {
					event.Old = oldVolume
				}
			}
			if rawEvent.New != nil {
				newVolume := &registry.Volume{}
				if err := volumeFromKeyValue(*rawEvent.New, newVolume); err != nil {
					event.Err = err
				} else {
					event.New = newVolume
				}
			}
			events <- event
		}
	}()

	return events
}

func (volRegistry *volumeRegistry) WaitForState(volumeName registry.VolumeName, state registry.VolumeState) error {
	log.Println("Start waiting for volume", volumeName, "to reach state", state)
	err := volRegistry.WaitForCondition(volumeName, func(event *registry.VolumeChange) bool {
		if event.New == nil {
			log.Panicf("unable to process event %+v", event)
		}
		return event.New.State == state || event.New.State == registry.Error
	})
	log.Println("Stopped waiting for volume", volumeName, "to reach state", state, err)
	if err != nil {
		return err
	}

	// return error if we went to an error state
	volume, err := volRegistry.Volume(volumeName)
	if err == nil && volume.State == registry.Error {
		return fmt.Errorf("stopped waiting as volume %s in error state", volumeName)
	}
	return err
}

// TODO: maybe have environment variable to tune this wait time?
var defaultTimeout = time.Minute * 10

func (volRegistry *volumeRegistry) WaitForCondition(volumeName registry.VolumeName,
	condition func(event *registry.VolumeChange) bool) error {

	volume, err := volRegistry.Volume(volumeName)
	if err != nil {
		return err
	}

	ctxt, cancelFunc := context.WithTimeout(context.Background(), defaultTimeout)
	events := volRegistry.GetVolumeChanges(ctxt, volume)
	defer cancelFunc()

	log.Printf("About to wait for condition on volume: %+v", volume)

	for event := range events {
		if event.Err != nil {
			return event.Err
		}
		if event.IsDelete {
			return fmt.Errorf("stopped waiting as volume %s is deleted", volume.Name)
		}

		conditionMet := condition(&event)
		if conditionMet {
			return nil
		}
	}

	return fmt.Errorf("stopped waiting for volume %s to meet supplied condition", volume.Name)
}
