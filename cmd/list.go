package cmd

import (
	"fmt"

	"github.com/swanwish/go-common/logs"
	"github.com/swanwish/godeps/models/godeps"
	"github.com/urfave/cli"
)

var (
	List = cli.Command{
		Name:        "list",
		Usage:       "List all the deps on godeps.json file",
		Description: "This command will list current deps",
		Action:      runList,
		Flags:       []cli.Flag{},
	}
)

func runList(c *cli.Context) error {
	goDeps, err := godeps.NewGoDeps()
	if err != nil {
		logs.Errorf("Failed to create godeps, the error is %v", err)
		return err
	}
	fmt.Println("The packages are:")
	for _, item := range goDeps.Deps {
		fmt.Printf("\npath: \t%s\norigin:\t%s\n", item.Path, item.Origin)
	}
	return nil
}
