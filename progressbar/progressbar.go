// progressbar is a package used for printing a progress-bar to an http.ResponseWriter
package progressbar

import (
	"fmt"
	"net/http"
)

type Bar struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	rate    string
	graph   string
	percent int
	current int
	total   int
}

// New is used to create a new ProgressBar.
func New(w http.ResponseWriter, f http.Flusher) *Bar {
	pb := &Bar{
		current: 0,
		total:   100,
		graph:   "█",
		writer:  w,
		flusher: f,
	}
	pb.init()

	return pb
}

// Update is called to update the progressbar progress.
func (b *Bar) Update(count int) {
	for i := 1; i <= 10; i++ {
		b.play(count + i)
	}
	b.flush()
	b.flusher.Flush()
}

func (b *Bar) init() {
	b.percent = b.getPercent()
	for i := 0; i < b.percent; i += 2 {
		b.rate += b.graph
	}
	b.Update(0)
}

func (b *Bar) getPercent() int {
	return int(float32(b.current) / float32(b.total) * 100)
}

func (b *Bar) play(cur int) {
	b.current = cur
	last := b.percent
	b.percent = b.getPercent()

	if b.percent != last && b.percent%2 == 0 {
		b.rate += b.graph
	}
}

func (b *Bar) flush() {
	fmt.Fprintf(b.writer, "\r[%-50s]%3d%% %8d/%d", b.rate, b.percent, b.current, b.total)
}
