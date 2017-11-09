package cmd

import (
	"github.com/swanwish/go-common/logs"
	"github.com/swanwish/godeps/models/godeps"
	"github.com/urfave/cli"
)

var (
	Init = cli.Command{
		Name:        "init",
		Usage:       "Create a godeps.json file in current folder",
		Description: "This command will create a godeps.json file in current folder",
		Action:      runInit,
		Flags:       []cli.Flag{},
	}
)

func runInit(c *cli.Context) error {
	goDeps, err := godeps.NewGoDeps()
	if err != nil {
		logs.Errorf("Failed to create godeps, the error is %v", err)
		return err
	}
	err = goDeps.Save()
	return err
}
