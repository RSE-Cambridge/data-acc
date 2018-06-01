package main

import (
	"fmt"
	"github.com/urfave/cli"
)

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

func getInstances() *instances {
	fakeInstance := instance{
		"fakebuffer",
		instanceCapacity{3, 40},
		instanceLinks{"fakebuffer"}}
	return &instances{fakeInstance}
}

func showInstances(_ *cli.Context) error {
	fmt.Print(getInstances())
	return nil
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

func getSessions() *sessions {
	fakeSession := session{"fakebuffer", 12345678, 1001, "fakebuffer"}
	return &sessions{fakeSession}
}

func showSessions(_ *cli.Context) error {
	fmt.Print(getSessions())
	return nil
}
