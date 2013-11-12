package monstrics

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/streadway/amqp"
	"io/ioutil"
	yaml "launchpad.net/goyaml"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Amqp    map[string]string `amqp,flow`
	ConfDir string            `confdir`
	Debug   bool              `debug`
	Actions []*Action         `,omitempty`
	Metrics []*Metric         `,omitempty`
	log     logging.Logger    `,omitempty`
}

func NewServer(filename string, log *logging.Logger) (*Server, error) {
	server := &Server{Actions: []*Action{}, Metrics: []*Metric{}}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return server, err
	}
	err = yaml.Unmarshal(content, &server)
	if err != nil {
		return server, err
	}

	confFiles, err := filepath.Glob(fmt.Sprintf("%s/*.yml", server.ConfDir))
	if err != nil || len(confFiles) == 0 {
		return server, fmt.Errorf("No metric or action files found, or error while reading them: %v", err)
	}

	for _, f := range confFiles {
		log.Info("Parsing file %s", f)
		err = parseActions(f, server)
		if err != nil {
			log.Warning("Error parsing actions: %s %v", f, err)
		}
		err = parseMetrics(f, server)
		if err != nil {
			log.Warning("Error parsing metrics: %s %v", f, err)
		}
	}

	// Convert the path lines to regexp matches
	for _, m := range server.Metrics {
		matchfrompath(m)
	}
	return server, nil
}

func matchfrompath(m *Metric) {
	quoted := regexp.QuoteMeta(m.Path)
	newMatch := regexp.MustCompile(`\\\*`).ReplaceAllString(quoted, `(.*)`)
	m.match = regexp.MustCompile(newMatch)
}

func (s *Server) String() string {
	r := fmt.Sprintf("\nAMQP: %v\n", s.Amqp)
	r += fmt.Sprintf("Conf dir: %v\n", s.ConfDir)
	for _, m := range s.Metrics {
		r += fmt.Sprintf("Metric:\n")
		r += fmt.Sprintf("Name :          %v\n", m.Name)
		r += fmt.Sprintf("Path :          %v\n", m.Path)
		r += fmt.Sprintf("Period:         %v\n", m.Period)
		r += fmt.Sprintf("duration:       %v\n", m.duration)
		r += fmt.Sprintf("Match:          %v\n", m.match)
		r += fmt.Sprintf("Transformation: %v\n", m.Transformations)
		r += fmt.Sprintf("Constraints:    %v\n", m.Constraints)
		r += fmt.Sprintf("Values:         %v\n", m.Values())
	}
	for _, a := range s.Actions {
		r += fmt.Sprintf("Action:\n")
		r += fmt.Sprintf("Action: %v\n", a.Action)
	}
	return r
}

// This setups the AMQP exchange and queue to receive data from clients (since the exchange
// may not exist)
func (s *Server) SetupAMQP() (c <-chan amqp.Delivery, conn *amqp.Connection, err error) {
	conn, err = amqp.Dial(s.Amqp["url"])
	if err != nil {
		return
	}
	s.log.Info("Connected to %s", s.Amqp["url"])
	ch, err := conn.Channel()
	if err != nil {
		return
	}
	// Declares the Exchange
	err = ch.ExchangeDeclare(s.Amqp["exchange"], "topic", true, false, false, false, nil)
	if err != nil {
		return
	}
	// Declares the Queue
	_, err = ch.QueueDeclare("monstrics_server", false, true, false, false, nil)
	if err != nil {
		return
	}
	err = ch.QueueBind("monstrics_server", "#", s.Amqp["exchange"], false, nil)
	if err != nil {
		return
	}
	s.log.Info("Sucessfully bound queue to exchange %s with pattern #", s.Amqp["exchange"])
	c, err = ch.Consume("monstrics_server", "", false, false, true, false, nil)
	if err != nil {
		return
	}
	return
}

func (s *Server) ProcessMessages(msg chan string, stop chan bool) {
	s.log.Debug("Starting message processor")
	for {
		select {
		case message := <-msg:
			s.log.Debug("Received on msg channel %s", message)
			lines := strings.Split(message, "\n")
			for _, line := range lines {
				r := regexp.MustCompile(`\s+`).Split(line, 3)
				path := r[0]
				value, err := strconv.ParseFloat(r[1], 64)
				if err != nil {
					s.log.Warning("Error %v", err)
				}
				ts, err := strconv.ParseFloat(r[2], 64)
				if err != nil {
					s.log.Warning("Error %v", err)
				}
				s.log.Debug("%v, %v, %v", path, value, ts)
				for _, m := range s.Metrics {
					if m.match.MatchString(path) {
						subs := m.match.FindStringSubmatch(path)
						// Interesting metric
						actual, exist := s.MetricbyPath(path)
						if !exist {
							// Metric path doesn't exist on the server yet
							newMetric := m.copy()
							if len(subs) == 2 {
								newMetric.Name = fmt.Sprintf(m.Name, subs[1])
							}
							newMetric.Path = path
							matchfrompath(newMetric)
							newMetric.SetValue(int64(ts), value)
							s.Metrics = append(s.Metrics, newMetric)
						} else {
							actual.SetValue(int64(ts), value)
							actual.trimValues()
						}
					}
				}
			}
		case <-stop:
			return
		}
	}
}

func (s *Server) MetricbyPath(path string) (*Metric, bool) {
	for _, m := range s.Metrics {
		m.RLock()
		if m.Path == path {
			m.RUnlock()
			return m, true
		}
		m.RUnlock()
	}
	return &Metric{}, false
}

func periodInDuration(period string) (duration time.Duration, err error) {
	var spec string
	match, err := regexp.Compile(`(\d+)(\S+)?`)
	if err != nil {
		return
	}
	subs := match.FindStringSubmatch(period)
	if len(subs) == 0 {
		err = fmt.Errorf("Badly formatted period")
		return
	} else if len(subs) == 3 {
		spec = subs[2]
	}
	value, err := strconv.Atoi(subs[1])
	if err != nil {
		return
	}
	switch spec {
	case "s", "":
		duration = time.Duration(value) * time.Second
	case "m":
		duration = time.Duration(value) * time.Minute
	case "h":
		duration = time.Duration(value) * time.Hour
	case "d":
		duration = time.Duration(value) * 24 * time.Hour
	}
	return
}

func parseActions(filename string, server *Server) (err error) {
	var action []*Action
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, &action)
	if err != nil {
		return err
	}
	for _, a := range action {
		if a.Action != "" {
			server.Actions = append(server.Actions, a)
		}
	}
	return nil
}

func parseMetrics(filename string, server *Server) (err error) {
	var metric []*Metric
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, &metric)
	if err != nil {
		return err
	}
	for _, m := range metric {
		if m.Path != "" {
			m.values = make(map[int64]float64)
			m.duration, err = periodInDuration(m.Period)
			if err != nil {
				return
			}
			server.Metrics = append(server.Metrics, m)
		}
	}
	return nil
}
