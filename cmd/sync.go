package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/swanwish/go-common/logs"
	"github.com/swanwish/go-common/utils"
	"github.com/swanwish/godeps/bash"
	"github.com/swanwish/godeps/models/godeps"
	"github.com/urfave/cli"
)

var (
	Sync = cli.Command{
		Name:        "sync",
		Usage:       "Sync packages from remote path",
		Description: "This commands will read the local godeps.json file, and sync the packages accordingly",
		Action:      runSync,
		Flags: []cli.Flag{
			stringFlag("path, p", "", "The path of the package"),
		},
	}
)

func runSync(c *cli.Context) error {
	path := c.String("path")
	if path != "" {
		logs.Debugf("Will update package with path is %s", path)
	}
	goDeps, err := godeps.NewGoDeps()
	if err != nil {
		logs.Errorf("Failed to get goDeps, the error is %v", err)
		return err
	}
	for _, item := range goDeps.Deps {
		if path != "" {
			if path == item.Path {
				return updateDepItem(item)
			}
		} else {
			if err = updateDepItem(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func updateDepItem(item *godeps.DepItem) error {
	logs.Debugf("Update package with origin %s", item.Origin)
	vendorPath := filepath.Join("vendor", item.Path)
	vendorParentPath := filepath.Dir(vendorPath)
	if !utils.FileExists(vendorParentPath) {
		err := os.MkdirAll(vendorParentPath, 0755)
		if err != nil {
			logs.Errorf("Failed to create dir with path %s, the error is %v", vendorParentPath, err)
			return err
		}
	}
	baseName := filepath.Base(vendorPath)
	command := ""
	if utils.FileExists(vendorPath) {
		logs.Debugf("Run git pull command")
		command = fmt.Sprintf("cd %s; git pull", vendorPath)
	} else {
		logs.Debugf("Run git clone command")
		command = fmt.Sprintf("cd %s; git clone %s %s", vendorParentPath, item.Origin, baseName)
	}
	output, err := bash.ExecuteCmd(command)
	if err != nil {
		logs.Errorf("Failed to fetch code from %s, the error is %v", item.Origin, err)
		return err
	}
	logs.Debugf("The output is %s", output)
	return nil
}
