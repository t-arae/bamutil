package main

import (
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
				Name:  "inbam",
				Value: "",
				Usage: "input bam file",
			},
			&cli.StringFlag{
				Name:  "outsam",
				Value: "-",
				Usage: "output sam file",
			},
			&cli.UintFlag{
				Name:    "num",
				Aliases: []string{"n"},
				Value:   10,
				Usage:   "number of output records",
			},
		},
		Action: func(c *cli.Context) error {
			num := c.Uint("num")
			br := openBamReader(c.String("inbam"))
			defer br.Close()
			h := getBamHeader(br)
			ch := make(chan *sam.Record)
			var wg sync.WaitGroup
			wg.Add(2)
			go sendBamRecord(br, num, ch, &wg)
			go writeRecordToSam(c.String("outsam"), h, ch, &wg)
			wg.Wait()
			return nil
		},
	}
}
