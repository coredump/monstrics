package monstrics

import (
	"fmt"
	"github.com/op/go-logging"
	"io/ioutil"
	yaml "launchpad.net/goyaml"
	"path"
	"path/filepath"
)

type Server struct {
	Amqp     map[string]string `amqp,flow`
	Conf_dir string            `conf_dir`
	Debug    bool              `debug`
	actions  []*Action         `,omitempty`
	metrics  []*Metric         `,omitempty`
	log      logging.Logger    `,omitempty`
}

func NewServer(filename string, log *logging.Logger) (*Server, error) {
	server := &Server{actions: []*Action{}, metrics: []*Metric{}}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return server, err
	}
	err = yaml.Unmarshal(content, &server)
	if err != nil {
		return server, err
	}

	actions_dir := path.Join(server.Conf_dir, "actions")
	metrics_dir := path.Join(server.Conf_dir, "metrics")
	action_files, err := filepath.Glob(fmt.Sprintf("%s/*.yml", actions_dir))
	if err != nil || len(action_files) == 0 {
		return server, fmt.Errorf("No action files found on %s, %v\n", actions_dir, err)
	}
	metric_files, err := filepath.Glob(fmt.Sprintf("%s/*.yml", metrics_dir))
	if err != nil || len(metric_files) == 0 {
		return server, fmt.Errorf("No metric files found on %s, %v\n", metrics_dir, err)
	}

	for _, f := range action_files {
		log.Info("Parsing file %s", f)
		err = parseActions(f, server)
		if err != nil {
			log.Warning("Error parsing actions: %s %v", f, err)
		}
	}
	for _, f := range metric_files {
		log.Info("Parsing file %s", f)
		err = parseMetrics(f, server)
		if err != nil {
			log.Warning("Error parsing metrics: %s %v", f, err)
		}
	}
	return server, nil
}

func parseActions(filename string, server *Server) (err error) {
	var action []map[string]*Action
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, &action)
	if err != nil {
		return err
	}
	for _, h := range action {
		server.actions = append(server.actions, h["handler"])
	}
	return nil
}

func parseMetrics(filename string, server *Server) (err error) {
	var metric []map[string]*Metric
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, &metric)
	if err != nil {
		return err
	}
	for _, m := range metric {
		server.metrics = append(server.metrics, m["metric"])
	}
	return nil
}
