package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"sync"

	"github.com/biogo/hts/sam"
	"github.com/urfave/cli/v2"
)

func cmdCountSoftClip() *cli.Command {
	return &cli.Command{
		Name:  "countsoftclip",
		Usage: "count removed soft clips from bam file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inbam",
				Aliases: []string{"i"},
				Value:   "",
				Usage:   "input bam file",
			},
			&cli.StringFlag{
				Name:    "outcsv",
				Aliases: []string{"o"},
				Value:   "",
				Usage:   "output csv file",
			},
		},
		Action: func(c *cli.Context) error {
			slog.Info("start countsoftclip")

			input := c.String("i")
			output := c.String("o")
			num := uint(0)

			br := openBamReader(input)
			defer br.Close()

			// Prepare sam.Record Stream
			ch := make(chan *sam.Record, 10000)

			var (
				tagLS = sam.NewTag("LS")
				tagRS = sam.NewTag("RS")
			)

			// Processing
			var wg sync.WaitGroup
			wg.Add(1)
			go sendBamRecord(br, num, ch, &wg)

			softclips_left := map[uint32]int{}
			softclips_right := map[uint32]int{}
			for rec := range ch {
				LS := rec.AuxFields.Get(tagLS)
				if LS != nil {
					lsv := LS.Value().(uint32)
					v, ok := softclips_left[lsv]
					if ok {
						softclips_left[lsv] = v + 1
					} else {
						softclips_left[lsv] = 1
					}
				}
				RS := rec.AuxFields.Get(tagRS)
				if RS != nil {
					rsv := RS.Value().(uint32)
					v, ok := softclips_right[rsv]
					if ok {
						softclips_right[rsv] = v + 1
					} else {
						softclips_right[rsv] = 1
					}
				}
			}
			wg.Wait()

			var w io.Writer
			switch output {
			case "", "-":
				w = os.Stdout
			default:
				f, err := os.Create(output)
				if err != nil {
					slog.Error(err.Error())
					os.Exit(1)
				}
				defer f.Close()
				w = f
			}
			wc := csv.NewWriter(w)
			wc.Write([]string{"side", "num_clips", "count"})
			for _, v := range getSortedKeys(softclips_left) {
				wc.Write([]string{"left", fmt.Sprintf("%d", v), fmt.Sprintf("%d", softclips_left[v])})
			}
			for _, v := range getSortedKeys(softclips_right) {
				wc.Write([]string{"right", fmt.Sprintf("%d", v), fmt.Sprintf("%d", softclips_right[v])})
			}
			wc.Flush()

			slog.Info("countsoftclip was succsessfully ended")
			return nil
		},
	}
}

func getSortedKeys(m map[uint32]int) []uint32 {
	keys := []uint32{}
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
