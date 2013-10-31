package monstrics

type Metric struct {
	Path        string            `path`
	Values      map[int64]float64 `values,omitempty`
	Constraints map[string]string `constraints`
}

type emailSettings struct {
}

type campfireSettings struct {
}

// Type Action receives the settings from all Actions decoded from the YAML files.
// It has to contains all possible fields from Actions, at least until goyaml supports
// decoding of embedded fields (then it can be made a little more cleaner)
type Action struct {
	Action    string   `action`
	Api_key   string   `api_key,omitempty`
	Rooms     []string `rooms,omitempty`
	Subdomain string   `subdomain,omitempty`
	Subject   string   `subject,omitempty`
	To        []string `to,omitempty`
}
