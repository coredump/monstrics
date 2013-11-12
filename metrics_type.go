package monstrics

import (
	"regexp"
	"sync"
	"time"
)

// Each received metric is converted to this
type Metric struct {
	sync.RWMutex
	Name            string            `name`
	Path            string            `path`
	values          map[int64]float64 `,omitempty`
	val_order       []int64           `,omitempty`
	Constraints     map[string]string `constraints`
	Transformations []string          `transformations,omitempty`
	match           *regexp.Regexp    `,omitempty`
	Period          string            `period`
	duration        time.Duration
}

func (m *Metric) copy() *Metric {
	new_metric := *m
	return &new_metric
}

func (m *Metric) SetValue(ts int64, val float64) {
	m.Lock()
	m.values[ts] = val
	m.val_order = append(m.val_order, ts)
	m.Unlock()
}

func (m *Metric) Values() (res map[int64]float64) {
	m.Lock()
	res = m.values
	m.Unlock()
	return
}

func (m *Metric) trimValues() {
	maxAge := time.Now().Unix() - int64(m.duration.Seconds())
	newOrder := make([]int64, len(m.val_order))
	m.Lock()
	for i, k := range m.val_order {
		if k <= maxAge {
			delete(m.values, k)
		} else {
			newOrder[i] = k
		}
	}
	m.val_order = newOrder
	m.Unlock()
}
