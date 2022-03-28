package main

import (
	"fmt"
	"net/http"
)

type Bar struct {
	Writer  *http.ResponseWriter
	percent int64
	cur     int64
	total   int64
	rate    string
	graph   string
}

func (bar *Bar) New(start, total int64) {
	bar.cur = start
	bar.total = total
	bar.graph = "#"
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph
	}
}

func (bar *Bar) getPercent() int64 {
	return int64(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *Bar) Play(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.graph
	}
}

func (bar *Bar) Flush() {
	_, _ = fmt.Fprintf(*bar.Writer, "\r[%-50s]%3d%% %8d/%d", bar.rate, bar.percent, bar.cur, bar.total)
}

func UpdateProgressBar(bar *Bar, count int64, w http.Flusher) {
	for i := 1; i <= 10; i++ {
		bar.Play(count + int64(i))
	}
	bar.Flush()
	w.Flush()

}
