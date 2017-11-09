package cmd

import (
	"github.com/swanwish/go-common/logs"
	"github.com/swanwish/godeps/common"
	"github.com/swanwish/godeps/models/godeps"
	"github.com/urfave/cli"
)

var (
	Update = cli.Command{
		Name:        "update",
		Usage:       "Update origin of the dep with path on godeps.json file",
		Description: "This command will update the dep's origin according to the path",
		Action:      runUpdate,
		Flags: []cli.Flag{
			stringFlag("path, p", "", "The path of the package"),
			stringFlag("origin, o", "", "The origin path of the git path"),
		},
	}
)

func runUpdate(c *cli.Context) error {
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
	return goDeps.UpdateItem(path, origin)
}
