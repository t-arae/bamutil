package main

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"math"
	"os"
	"sync"

	"github.com/biogo/hts/bam"
	"github.com/biogo/hts/bgzf"
	"github.com/biogo/hts/sam"
)

func openBamReader(inf string) *bam.Reader {
	var r io.Reader
	if inf == "" {
		r = os.Stdin
	} else if inf == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(inf)
		if err != nil {
			slog.Error(fmt.Sprintf("could not open the input file %q\n", inf))
			os.Exit(1)
		}
		ok, err := bgzf.HasEOF(f)
		if err != nil {
			slog.Error(err.Error())
			slog.Error(fmt.Sprintf("invalid EOF in %q\n", inf))
			os.Exit(1)
		}
		if !ok {
			slog.Error(fmt.Sprintf("file %q has no bgzf magic block", inf))
			os.Exit(1)
		}
		r = f
	}

	br, err := bam.NewReader(r, 0)
	if err != nil {
		slog.Error(err.Error())
	}

	return br
}

func getBamHeader(br *bam.Reader) *sam.Header {
	h := br.Header().Clone()
	return h
}

func sendBamRecord(br *bam.Reader, num uint, ch chan *sam.Record, wg *sync.WaitGroup) {
	if num == 0 {
		num = math.MaxUint
	}
	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func(ch chan<- *sam.Record, wg2 *sync.WaitGroup) {
		for i := uint(0); i < num; i++ {
			rec, err := br.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
			ch <- rec
		}
		wg2.Done()
	}(ch, &wg2)
	wg2.Wait()
	close(ch)
	wg.Done()
}

func writeRecordToBam(outf string, header *sam.Header, ch chan *sam.Record, wg *sync.WaitGroup) {
	var w io.Writer
	var err error
	if outf == "" {
		slog.Error("no output file specified\n")
	} else if outf == "-" {
		w = os.Stdout
	} else {
		w, err = os.Create(outf)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
	}
	bw, err := bam.NewWriter(w, header, 0)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func(bw *bam.Writer, ch chan *sam.Record) {
		for rec := range ch {
			err = bw.Write(rec)
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
		}
		wg2.Done()
	}(bw, ch)
	wg2.Wait()
	wg.Done()
}

func writeRecordToSam(outf string, header *sam.Header, ch chan *sam.Record, wg *sync.WaitGroup) {
	var w io.Writer
	var err error
	if outf == "" {
		slog.Error("no output file specified\n")
	} else if outf == "-" {
		w = os.Stdout
	} else {
		w, err = os.Create(outf)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		//defer w.Close()
	}
	bw := bufio.NewWriter(w)
	sw, err := sam.NewWriter(bw, header, 0)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func(sw *sam.Writer, ch chan *sam.Record) {
		for rec := range ch {
			err = sw.Write(rec)
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
		}
		wg2.Done()
	}(sw, ch)
	wg2.Wait()
	wg.Done()
}
