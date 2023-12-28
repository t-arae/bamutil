package main

import (
	"log/slog"
	"sync"

	"github.com/biogo/hts/sam"
	"github.com/urfave/cli/v2"
)

func cmdView() *cli.Command {
	return &cli.Command{
		Name:  "bam2sam",
		Usage: "view bam file as sam file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inbam",
				Aliases: []string{"i"},
				Value:   "",
				Usage:   "input bam file",
			},
			&cli.StringFlag{
				Name:    "outsam",
				Aliases: []string{"o"},
				Value:   "-",
				Usage:   "output sam file",
			},
			&cli.UintFlag{
				Name:    "num",
				Aliases: []string{"n"},
				Value:   10,
				Usage:   "number of output records",
			},
		},
		Action: func(c *cli.Context) error {
			slog.Info("start bam2sam")

			input := c.String("i")
			output := c.String("o")
			num := c.Uint("num")

			br := openBamReader(input)
			defer br.Close()
			h := getBamHeader(br)

			// Prepare sam.Record Stream
			ch := make(chan *sam.Record)

			// Processing
			var wg sync.WaitGroup
			wg.Add(2)
			go sendBamRecord(br, num, ch, &wg)
			go writeRecordToSam(output, h, ch, &wg)
			wg.Wait()
			slog.Info("bam2sam was succsessfully ended")
			return nil
		},
	}
}
