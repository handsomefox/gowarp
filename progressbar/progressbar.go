// progressbar is a package used for printing a progress-bar to an http.ResponseWriter
package progressbar

import (
	"fmt"
	"net/http"
)

const (
	graph  = "█"                     // used to show the progress
	format = "\r[%-50s]%3d%% %8d/%d" // progress bar format
)

type Bar struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	rate    string
	percent int
	current int
	total   int
}

// New is used to create a new ProgressBar.
func New(w http.ResponseWriter, f http.Flusher) *Bar {
	pb := &Bar{
		current: 0,
		total:   100,
		writer:  w,
		flusher: f,
	}
	pb.init()

	return pb
}

// Update is called to update the progressbar progress.
func (bar *Bar) Update(count int) {
	for i := 1; i <= 10; i++ {
		bar.play(count + i)
	}
	bar.flush()
	bar.flusher.Flush()
}

func (bar *Bar) init() {
	bar.percent = bar.getPercent()
	for i := 0; i < bar.percent; i += 2 {
		bar.rate += graph
	}
	bar.Update(0)
}

func (bar *Bar) getPercent() int {
	return int(float32(bar.current) / float32(bar.total) * 100)
}

func (bar *Bar) play(cur int) {
	bar.current = cur
	last := bar.percent
	bar.percent = bar.getPercent()

	if bar.percent != last && bar.percent%2 == 0 {
		bar.rate += graph
	}
}

func (bar *Bar) flush() {
	fmt.Fprintf(bar.writer, format, bar.rate, bar.percent, bar.current, bar.total)
}
