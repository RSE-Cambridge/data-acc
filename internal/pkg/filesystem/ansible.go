package filesystem

import "github.com/RSE-Cambridge/data-acc/internal/pkg/datamodel"

type AnsibleEnv struct {
	directory string
}

type Ansible interface {
	CreateEnvironment(session datamodel.Session, createAllPlaybooks bool, groupAllVars map[string]string) (AnsibleEnv, error)
	DestroyEnvironment(env AnsibleEnv) error
	RunPlaybook(env AnsibleEnv, playbook AnsiblePlaybook, extraVars map[string]string)
}

type AnsiblePlaybook int

const (
	Create AnsiblePlaybook = iota
	Delete
	DataCopyIn
	DataCopyOut
	// Mount and unmount require extraVars with the hosts to be mounted
	Mount
	Unmount
)
