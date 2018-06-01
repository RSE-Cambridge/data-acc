package main

import (
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/fakewarp"
	"github.com/urfave/cli"
)

func showInstances(_ *cli.Context) error {
	fmt.Print(fakewarp.GetInstances())
	return nil
}

func showSessions(_ *cli.Context) error {
	fmt.Print(fakewarp.GetSessions())
	return nil
}

func listPools(_ *cli.Context) error {
	fmt.Print(fakewarp.GetPools())
	return nil
}
