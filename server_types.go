package monstrics

type Metric struct {
	Path        string            `path`
	Values      map[int64]float64 `values,omitempty`
	Constraints map[string]string `constraints`
}

type Action struct {
	Action string            `action`
	Vars   map[string]string `vars`
}
