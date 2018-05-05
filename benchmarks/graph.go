package main

import (
	"fmt"
	"math"
	"os/exec"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

type Metric struct {
	param int64
	t     time.Duration
}

type Metrics []Metric

func (ms Metrics) Len() int {
	return len(ms)
}

func (ms Metrics) XY(i int) (float64, float64) {
	return float64(ms[i].param), float64(ms[i].t.Nanoseconds())
}

type PerfPlot struct {
	seriesOrder []string
	series      map[string]Metrics
}

func NewPerfPlot() *PerfPlot {
	p := new(PerfPlot)
	p.Clear()
	return p
}

func (p *PerfPlot) Clear() {
	p.seriesOrder = make([]string, 0)
	p.series = make(map[string]Metrics)
}

func (p *PerfPlot) AddMetric(series string, param int64, t time.Duration) {
	var s []Metric
	var ok bool

	if s, ok = p.series[series]; !ok {
		p.seriesOrder = append(p.seriesOrder, series)
		s = make([]Metric, 0, 100)
	}

	s = append(s, Metric{param, t})

	p.series[series] = s
}

func (p *PerfPlot) Plot(file, xlabel, ylabel, title string) {

	seriesargs := make([]interface{}, 0, len(p.series)*2)

	pl, err := plot.New()
	if err != nil {
		panic(err)
	}

	pl.Title.Text = title
	pl.X.Label.Text = xlabel
	pl.Y.Label.Text = ylabel
	pl.X.Tick.Marker = pwr2Ticks{}
	pl.X.Tick.Label.Rotation = -math.Pi / 2
	pl.X.Tick.Label.Font.Size = vg.Millimeter * 3
	pl.X.Scale = plot.LogScale{}
	pl.Legend.Top = true
	pl.Add(plotter.NewGrid())

	for _, s := range p.seriesOrder {
		v := p.series[s]
		seriesargs = append(seriesargs, s, v)
	}

	err = plotutil.AddLinePoints(pl, seriesargs...)
	if err != nil {
		panic(err)
	}

	if err := pl.Save(8*vg.Inch, 6*vg.Inch, file); err != nil {
		panic(err)
	}
}

type pwr2Ticks struct{}

func (pwr2Ticks) Ticks(min, max float64) []plot.Tick {
	tks := make([]plot.Tick, 0)

	for i := min; i <= max; i *= 2 {
		tks = append(tks, plot.Tick{
			Label: fmt.Sprint(int64(i)),
			Value: i,
		})
	}

	return tks
}

// AddCommas adds commas after every 3 characters from right to left.
// NOTE: This function is a quick hack, it doesn't work with decimal
// points, and may have a bunch of other problems.
func addCommas(s string) string {
	rev := ""
	n := 0
	for i := len(s) - 1; i >= 0; i-- {
		rev += string(s[i])
		n++
		if n%3 == 0 {
			rev += ","
		}
	}
	s = ""
	for i := len(rev) - 1; i >= 0; i-- {
		s += string(rev[i])
	}
	return s
}

// Returns err if we could not open viewer
func OpenPlot(filename string) error {
	const viewer = "eog"
	_, err := exec.LookPath(viewer)
	if err != nil {
		return err
	}

	cmd := exec.Command(viewer, filename)
	err = cmd.Start()
	// cmd.Process.Release()
	return err
}
