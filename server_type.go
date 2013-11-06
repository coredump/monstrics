package monstrics

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/streadway/amqp"
	"io/ioutil"
	yaml "launchpad.net/goyaml"
	"path/filepath"
)

type Server struct {
	Amqp     map[string]string `amqp,flow`
	Conf_dir string            `conf_dir`
	Debug    bool              `debug`
	Actions  []*Action         `,omitempty`
	Metrics  []*Metric         `,omitempty`
	log      logging.Logger    `,omitempty`
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

	conf_files, err := filepath.Glob(fmt.Sprintf("%s/*.yml", server.Conf_dir))
	if err != nil || len(conf_files) == 0 {
		return server, fmt.Errorf("No metric or action files found, or error while reading them: %v", err)
	}

	for _, f := range conf_files {
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

	return server, nil
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
			server.Metrics = append(server.Metrics, m)
		}
	}
	return nil
}
