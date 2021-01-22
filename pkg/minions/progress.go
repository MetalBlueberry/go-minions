package minions

import "sync"

// Progress provides an easy way to report progress.
// The progress is calculated as the current progress over the total progress
type Progress struct {
	report  string
	total   int
	current int

	m sync.Mutex
}

// Add increases the total count
func (p *Progress) Add(n int) {
	p.m.Lock()
	defer p.m.Unlock()
	p.total += n
}

// SetReport allows to store a message to be displayed with the progress report
func (p *Progress) SetReport(msg string) {
	p.m.Lock()
	defer p.m.Unlock()
	p.report = msg
}

// Done increases in 1 the current progress
func (p *Progress) Done() {
	p.m.Lock()
	defer p.m.Unlock()
	p.current++
}

// DoneN increases in n the current progress
func (p *Progress) DoneN(n int) {
	p.m.Lock()
	defer p.m.Unlock()
	p.current += n
}

// Report returns the string report
func (p *Progress) Report() string {
	p.m.Lock()
	defer p.m.Unlock()
	return p.report
}

// Percentage returns the % current/total
func (p *Progress) Percentage() float64 {
	p.m.Lock()
	defer p.m.Unlock()
	return 100.0 * float64(p.current) / float64(p.total)
}

// Progress returns the individual values of current and total
func (p *Progress) Progress() (current int, total int) {
	p.m.Lock()
	defer p.m.Unlock()
	return p.current, p.total
}
