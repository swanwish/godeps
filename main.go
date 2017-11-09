package main

import (
	"os"
	"runtime"

	"github.com/swanwish/godeps/cmd"
	"github.com/urfave/cli"
)

const APP_VERSION = "0.1.0"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	app := cli.NewApp()
	app.Name = "godeps"
	app.Usage = "The tool for manage the dependent packages for go project"
	app.Version = APP_VERSION
	app.Commands = []cli.Command{
		cmd.Add,
		cmd.Delete,
		cmd.Update,
		cmd.List,
		cmd.Sync,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)
	app.Run(os.Args)
}
