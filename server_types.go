package monstrics

type Constraint struct {
	Name  string
	Value string
}

type Metric struct {
	Path        string
	Values      map[int64]float64
	Constraints []*Constraint
}

type Handler interface {
	Handle()
}

type Action struct {
	Name        string
	Description string
	Vars        map[string]string
}

type ServerConfigFile struct {
	Amqp     map[string]string `amqp,flow`
	Conf_dir string            `conf_dir`
	Debug    bool              `debug`
}

type Server struct {
	Handlers []Handler
	Metrics  []Metric
	Checkers []func() // TODO
}
