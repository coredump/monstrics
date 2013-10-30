package monstrics

import (
	"io/ioutil"
	yaml "launchpad.net/goyaml"
)

func NewServerConfig(filename string) (err error, config *ServerConfigFile) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err, config
	}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		return err, config
	}
	return nil, config
}
