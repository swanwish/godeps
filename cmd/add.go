package cmd

import (
	"github.com/swanwish/go-common/logs"
	"github.com/swanwish/godeps/common"
	"github.com/swanwish/godeps/models/godeps"
	"github.com/urfave/cli"
)

var (
	Add = cli.Command{
		Name:        "add",
		Usage:       "Add one deps on godeps.json file",
		Description: "This command will add the deps to the project",
		Action:      runAdd,
		Flags: []cli.Flag{
			stringFlag("path, p", "", "The path of the package"),
			stringFlag("origin, o", "", "The origin path of the git path"),
		},
	}
)

func runAdd(c *cli.Context) error {
	path := c.String("path")
	origin := c.String("origin")
	if path == "" || origin == "" {
		logs.Errorf("The path or origin is not specified")
		return common.ErrInvalidParameter
	}
	goDeps, err := godeps.NewGoDeps()
	if err != nil {
		logs.Errorf("Failed to create godeps, the error is %v", err)
		return err
	}

	if err = goDeps.AddItem(path, origin); err != nil {
		return err
	}

	return goDeps.Save()
}
