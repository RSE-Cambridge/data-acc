package main

import "github.com/urfave/cli"

type pool struct {
	Id          string `json:"id"`
	Units       string `json:"units"`
	Granularity uint   `json:"granularity"`
	Quantity    uint   `json:"quantity"`
	Free        uint   `json:"free"`
}

func pools(_ *cli.Context) error {
	p := pool{"fake", "bytes", 214748364800, 40, 3}
	m := map[string]pool{"pool": p}
	printJson(m)
	return nil
}
