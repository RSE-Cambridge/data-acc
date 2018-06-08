package fakewarp

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

func GetInstances(volRegistry registry.VolumeRegistry) (*instances, error) {
	fakeInstance := instance{
		"fakebuffer",
		instanceCapacity{3, 40},
		instanceLinks{"fakebuffer"}}
	return &instances{fakeInstance}, nil
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
	fakeSession := session{"fakebuffer", 1234567890, 1001, "fakebuffer"}
	return &sessions{fakeSession}, nil
}

type configurations []string

func (list *configurations) String() string {
	message := map[string]configurations{"configurations": *list}
	return toJson(message)
}

func GetConfigurations() *configurations {
	return &configurations{}
}
