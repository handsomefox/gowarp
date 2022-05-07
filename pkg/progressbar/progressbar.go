package progressbar

import (
	"fmt"
	"net/http"
)

type Progressbar struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	percent int
	current int
	total   int
	rate    string
	graph   string
}

func New(w http.ResponseWriter, f http.Flusher) *Progressbar {
	pb := &Progressbar{
		current: 0,
		total:   100,
		graph:   "█",
		writer:  w,
		flusher: f,
	}
	pb.init()
	return pb
}

func (b *Progressbar) Update(count int) {
	for i := 1; i <= 10; i++ {
		b.play(count + i)
	}
	b.flush()
	b.flusher.Flush()
}

func (b *Progressbar) init() {
	b.percent = b.getPercent()
	for i := 0; i < b.percent; i += 2 {
		b.rate += b.graph
	}
	b.Update(0)
}

func (b *Progressbar) getPercent() int {
	return int(float32(b.current) / float32(b.total) * 100)
}

func (b *Progressbar) play(cur int) {
	b.current = cur
	last := b.percent
	b.percent = b.getPercent()
	if b.percent != last && b.percent%2 == 0 {
		b.rate += b.graph
	}
}

func (b *Progressbar) flush() {
	fmt.Fprintf(b.writer, "\r[%-50s]%3d%% %8d/%d", b.rate, b.percent, b.current, b.total)
}
