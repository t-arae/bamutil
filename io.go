package main

import (
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
	slog.Info(fmt.Sprintf("input file: %#q\n", inf))
	var r io.Reader
	if inf == "" || inf == "-" {
		slog.Info("read from `stdin`\n")
		r = os.Stdin
	} else {
		f, err := os.Open(inf)
		if err != nil {
			slog.Error(fmt.Sprintf("could not open the input file %#q\n", inf))
			os.Exit(1)
		}
		ok, err := bgzf.HasEOF(f)
		if err != nil {
			slog.Error(fmt.Sprintf("invalid EOF in %#q\n", inf))
			os.Exit(1)
		}
		if !ok {
			slog.Error(fmt.Sprintf("file %#q has no bgzf magic block", inf))
			os.Exit(1)
		}
		r = f
	}

	br, err := bam.NewReader(r, 0)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return br
}

func getBamHeader(br *bam.Reader) *sam.Header {
	h := br.Header().Clone()
	return h
}

func sendBamRecord(br *bam.Reader, num uint, ch chan *sam.Record, wg *sync.WaitGroup) {
	if num == 0 {
		slog.Info("all reads will be used\n")
		num = math.MaxUint
	} else {
		slog.Info(fmt.Sprintf("%d reads will be used\n", num))
	}

	var wg2 sync.WaitGroup
	wg2.Add(1)
	n := 0
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
			n++
		}
		wg2.Done()
	}(ch, &wg2)
	wg2.Wait()
	slog.Info(fmt.Sprintf("read %d records\n", n))
	close(ch)
	wg.Done()
}

func writeRecordToBam(outf string, header *sam.Header, ch chan *sam.Record, wg *sync.WaitGroup) {
	slog.Info(fmt.Sprintf("output file: %#q\n", outf))
	var w io.Writer
	var err error
	if outf == "" || outf == "-" {
		slog.Info("write to `stdout`\n")
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
	n := 0
	go func(bw *bam.Writer, ch chan *sam.Record) {
		for rec := range ch {
			err = bw.Write(rec)
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
			n++
		}
		wg2.Done()
	}(bw, ch)
	wg2.Wait()
	bw.Close()
	slog.Info(fmt.Sprintf("wrote %v records\n", n))
	wg.Done()
}

func writeRecordToSam(outf string, header *sam.Header, ch chan *sam.Record, wg *sync.WaitGroup) {
	slog.Info(fmt.Sprintf("output file: %#q\n", outf))
	var w io.Writer
	var err error
	if outf == "" || outf == "-" {
		slog.Info("write to `stdout`\n")
		w = os.Stdout
	} else {
		w, err = os.Create(outf)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		//defer w.Close()
	}
	sw, err := sam.NewWriter(w, header, 0)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	var wg2 sync.WaitGroup
	wg2.Add(1)
	n := 0
	go func(sw *sam.Writer, ch chan *sam.Record) {
		for rec := range ch {
			err = sw.Write(rec)
			if err != nil {
				slog.Error(err.Error())
				os.Exit(1)
			}
			n++
		}
		wg2.Done()
	}(sw, ch)
	wg2.Wait()
	slog.Info(fmt.Sprintf("wrote %v records\n", n))
	wg.Done()
}
