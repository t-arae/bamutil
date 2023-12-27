package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "bamutil"
	app.Usage = "Miscellaneous BAM/SAM utilities"
	app.Version = "0.0.1"

	app.Commands = []*cli.Command{
		cmdRemoveSoftClip(),
		cmdView(),
	}

	app.Run(os.Args)
}
