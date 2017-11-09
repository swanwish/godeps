package cmd

import (
	"github.com/swanwish/go-common/logs"
	"github.com/swanwish/godeps/common"
	"github.com/swanwish/godeps/models/godeps"
	"github.com/urfave/cli"
)

var (
	Delete = cli.Command{
		Name:        "delete",
		Usage:       "Delete one dep from godeps.json file",
		Description: "This command will delete a dep from godeps.json file with the path",
		Action:      runDelete,
		Flags: []cli.Flag{
			stringFlag("path, p", "", "The path of the package"),
		},
	}
)

func runDelete(c *cli.Context) error {
	path := c.String("path")
	if path == "" {
		logs.Errorf("The path is not specified")
		return common.ErrInvalidParameter
	}
	goDeps, err := godeps.NewGoDeps()
	if err != nil {
		logs.Errorf("Failed to create godeps, the error is %v", err)
		return err
	}
	err = goDeps.DeleteItem(path)
	return err
}
