package cli

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type pullProgressCounts struct {
	DriveDiscovered int
	WikiDiscovered  int
	DriveExported   int
	WikiExported    int
	Errors          int
}

type pullProgress struct {
	out     io.Writer
	style   termStyle
	interval time.Duration

	sync.Mutex
	stage        string
	counts       pullProgressCounts
	lastChange   time.Time
	lastPrinted  time.Time
	closed       bool
	stopCh       chan struct{}
}

func newPullProgress(out io.Writer, interval time.Duration) *pullProgress {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	now := time.Now()
	p := &pullProgress{
		out:        out,
		style:      newTermStyle(out),
		interval:   interval,
		lastChange: now,
		stopCh:     make(chan struct{}),
	}
	go p.loop()
	return p
}

func (p *pullProgress) Close() {
	p.Lock()
	if p.closed {
		p.Unlock()
		return
	}
	p.closed = true
	close(p.stopCh)
	p.Unlock()
}

func (p *pullProgress) SetStage(stage string) {
	p.Lock()
	changed := p.stage != stage
	p.stage = stage
	if changed {
		p.lastChange = time.Now()
	}
	p.Unlock()
	if changed {
		p.print(true)
	}
}

func (p *pullProgress) AddDriveDiscovered(n int) { p.bump(func(c *pullProgressCounts) { c.DriveDiscovered += n }) }
func (p *pullProgress) AddWikiDiscovered(n int)  { p.bump(func(c *pullProgressCounts) { c.WikiDiscovered += n }) }
func (p *pullProgress) AddDriveExported(n int)   { p.bump(func(c *pullProgressCounts) { c.DriveExported += n }) }
func (p *pullProgress) AddWikiExported(n int)    { p.bump(func(c *pullProgressCounts) { c.WikiExported += n }) }
func (p *pullProgress) AddErrors(n int)          { p.bump(func(c *pullProgressCounts) { c.Errors += n }) }

func (p *pullProgress) bump(fn func(*pullProgressCounts)) {
	p.Lock()
	before := p.counts
	fn(&p.counts)
	after := p.counts
	changed := before != after
	if changed {
		p.lastChange = time.Now()
	}
	p.Unlock()
	if changed {
		p.print(false)
	}
}

func (p *pullProgress) loop() {
	t := time.NewTicker(p.interval)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			p.print(false)
		case <-p.stopCh:
			return
		}
	}
}

func (p *pullProgress) print(force bool) {
	p.Lock()
	stage := p.stage
	c := p.counts
	lastChange := p.lastChange
	now := time.Now()
	should := force || p.lastPrinted.IsZero() || now.Sub(p.lastPrinted) >= p.interval
	if !should {
		p.Unlock()
		return
	}
	p.lastPrinted = now
	p.Unlock()

	stageStr := stage
	if stageStr == "" {
		stageStr = "running"
	}

	stale := now.Sub(lastChange)
	still := ""
	if stale >= p.interval {
		still = " " + p.style.faint(fmt.Sprintf("still working (last +%s)", stale.Round(time.Second)))
	}

	fmt.Fprintf(p.out, "%s %s drive=%s wiki=%s exported=%s/%s errors=%s%s\n",
		p.style.heading("[pull]"),
		p.style.bold(stageStr),
		p.style.bold(fmt.Sprintf("%d", c.DriveDiscovered)),
		p.style.bold(fmt.Sprintf("%d", c.WikiDiscovered)),
		p.style.bold(fmt.Sprintf("%d", c.DriveExported)),
		p.style.bold(fmt.Sprintf("%d", c.WikiExported)),
		p.style.warn(fmt.Sprintf("%d", c.Errors)),
		still,
	)
}
