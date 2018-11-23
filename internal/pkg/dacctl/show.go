package dacctl

import "github.com/RSE-Cambridge/data-acc/internal/pkg/registry"

type instanceCapacity struct {
	Bytes uint `json:"bytes"`
	Nodes uint `json:"nodes"`
}

type instanceLinks struct {
	Session string `json:"session"`
}

type instance struct {
	Id       string           `json:"id"`
	Capacity instanceCapacity `json:"capacity"`
	Links    instanceLinks    `json:"links"`
}

type instances []instance

func (list *instances) String() string {
	message := map[string]instances{"instances": *list}
	return toJson(message)
}

const bytesInGB = 1073741824

func GetInstances(volRegistry registry.VolumeRegistry) (*instances, error) {
	instances := instances{}
	volumes, err := volRegistry.AllVolumes()
	if err != nil {
		// TODO... normally means there are no instances
	}

	for _, volume := range volumes {
		instances = append(instances, instance{
			Id:       string(volume.Name),
			Capacity: instanceCapacity{Bytes: volume.SizeGB * bytesInGB, Nodes: volume.SizeBricks},
			Links:    instanceLinks{volume.JobName},
		})
	}
	return &instances, nil
}

type session struct {
	Id      string `json:"id"`
	Created uint   `json:"created"`
	Owner   uint   `json:"owner"`
	Token   string `json:"token"`
}

type sessions []session

func (list *sessions) String() string {
	message := map[string]sessions{"sessions": *list}
	return toJson(message)
}

func GetSessions(volRegistry registry.VolumeRegistry) (*sessions, error) {
	jobs, err := volRegistry.Jobs()
	if err != nil {
		// TODO: error is usually about there not being any jobs
		jobs = []registry.Job{}
	}
	sessions := sessions{}
	for _, job := range jobs {
		sessions = append(sessions, session{
			Id:      job.Name,
			Created: job.CreatedAt,
			Owner:   job.Owner,
			Token:   job.Name,
		})
	}
	return &sessions, nil
}

type configurations []string

func (list *configurations) String() string {
	message := map[string]configurations{"configurations": *list}
	return toJson(message)
}

func GetConfigurations() *configurations {
	return &configurations{}
}
