package progressbar

import (
	"fmt"
	"net/http"
)

type Progressbar struct {
	Writer  http.ResponseWriter
	Flusher http.Flusher

	percent int64
	cur     int64
	total   int64
	rate    string
	graph   string
}

func New(w http.ResponseWriter, f http.Flusher) *Progressbar {
	pb := &Progressbar{
		Writer:  w,
		Flusher: f,
	}
	pb.init(0, 100)
	pb.Update(0)
	return pb
}

func (bar *Progressbar) Update(count int64) {
	for i := 1; i <= 10; i++ {
		bar.play(count + int64(i))
	}
	bar.flush()
	bar.Flusher.Flush()
}

func (bar *Progressbar) init(start, total int64) {
	bar.cur = start
	bar.total = total
	bar.graph = "█"
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph
	}
}

func (bar *Progressbar) getPercent() int64 {
	return int64(float32(bar.cur) / float32(bar.total) * 100)
}

func (bar *Progressbar) play(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += bar.graph
	}
}

func (bar *Progressbar) flush() {
	_, _ = fmt.Fprintf(bar.Writer, "\r[%-50s]%3d%% %8d/%d", bar.rate, bar.percent, bar.cur, bar.total)
}
