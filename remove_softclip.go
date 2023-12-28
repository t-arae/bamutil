package main

import (
	"log/slog"
	"os"
	"sync"

	"github.com/biogo/hts/sam"
	"github.com/urfave/cli/v2"
)

func cmdRemoveSoftClip() *cli.Command {
	return &cli.Command{
		Name:  "rmsoftclip",
		Usage: "remove soft clips from bam file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "inbam",
				Aliases: []string{"i"},
				Value:   "",
				Usage:   "input bam file",
			},
			&cli.StringFlag{
				Name:    "outbam",
				Aliases: []string{"o"},
				Value:   "",
				Usage:   "output bam file",
			},
		},
		Action: func(c *cli.Context) error {
			slog.Info("start rmsoftclip")

			input := c.String("i")
			output := c.String("o")
			num := uint(0)

			br := openBamReader(input)
			defer br.Close()
			h := getBamHeader(br)

			// Prepare sam.Record Stream
			ch := make(chan *sam.Record, 10000)
			chRmSoftClip := make(chan *sam.Record, 10000)

			// Processing
			var wg sync.WaitGroup
			wg.Add(3)
			go sendBamRecord(br, num, ch, &wg)
			go func(chIn chan *sam.Record, chOut chan *sam.Record, wg *sync.WaitGroup) {
				for rec := range chIn {
					chOut <- removeSoftClip(rec)
				}
				wg.Done()
				close(chOut)
			}(ch, chRmSoftClip, &wg)
			go writeRecordToBam(output, h, chRmSoftClip, &wg)
			wg.Wait()
			slog.Info("rmsoftclip was succsessfully ended")
			return nil
		},
	}
}

var (
	tagLS = sam.NewTag("LS")
	tagRS = sam.NewTag("RS")
)

func removeSoftClip(rec *sam.Record) *sam.Record {
	cigar := rec.Cigar

	// Extract left and right soft clips
	var ls, rs int
	if cigar[0].Type() == sam.CigarSoftClipped {
		ls = cigar[0].Len()
		cigar = cigar[1:]
	} else {
		ls = 0
	}
	len_cigar := len(cigar)
	if cigar[len_cigar-1].Type() == sam.CigarSoftClipped {
		rs = cigar[len_cigar-1].Len()
		cigar = cigar[0:(len_cigar - 1)]
	} else {
		rs = 0
	}
	auxLS, err := sam.NewAux(tagLS, uint32(ls))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	auxRS, err := sam.NewAux(tagRS, uint32(rs))
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	rec.Seq = sam.NewSeq(rec.Seq.Expand()[ls:(rec.Seq.Length - rs)])
	rec.Qual = rec.Qual[ls:(len(rec.Qual) - rs)]
	rec.Cigar = cigar
	rec.AuxFields = append(rec.AuxFields, auxLS, auxRS)

	return rec
}
