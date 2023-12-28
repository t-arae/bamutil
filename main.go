package main

import (
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "bamutil"
	app.Usage = "Miscellaneous BAM/SAM utilities"
	app.Version = "0.0.2"

	var logger *slog.Logger
	logger = slog.Default()
	slog.SetDefault(logger)

	app.Commands = []*cli.Command{
		cmdRemoveSoftClip(),
		cmdCountSoftClip(),
		cmdView(),
	}

	app.Run(os.Args)

}
