package monstrics

import (
	"fmt"
	"github.com/op/go-logging"
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
