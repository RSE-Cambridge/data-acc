package filesystem_impl

import (
	"bytes"
	"encoding/json"
)

type FSType int

const (
	BeegFS FSType = iota
	Lustre
)

var fsTypeStrings = map[FSType]string{
	BeegFS: "BeegFS",
	Lustre: "Lustre",
}
var stringToFSType = map[string]FSType{
	"":       BeegFS,
	"BeegFS": BeegFS,
	"Lustre": Lustre,
}

func (fsType FSType) String() string {
	return fsTypeStrings[fsType]
}

func (fsType FSType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(fsTypeStrings[fsType])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (fsType *FSType) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return err
	}
	*fsType = stringToFSType[str]
	return nil
}
