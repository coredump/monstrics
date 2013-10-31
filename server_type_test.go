package monstrics

import (
	"fmt"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"testing"
)

var (
	action_yml = `---
- handler:
    action: campfire
    vars:
      rooms: "Operations Room,Ops Tears"
      campfire_api_key: "asdkAJDKJ#K@JK#JDK@J#DKJ@"
      campfire_subdomain: "testing"
`
	metric_yml = `---
- metric:
    path: stats.production.*.unicorn.socket_queued
    constraints:
      Upper Limit: 5
      Lower Limit: 5
`
	conf_yml = `---
amqp:
  url: amqp://guest:guest@localhost/
  exchange: metrics
conf_dir: /xxxxx
debug: true
`
	file_prefix = "go_test_monstrics"
)

func TestNewServer(t *testing.T) {
	correct_conf, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = correct_conf.WriteString(conf_yml)
	wrong_conf, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = wrong_conf.WriteString("---\n{bogus\n:yml")
	if err != nil {
		t.Fatal("Failed to setup conf file")
	}
	defer os.Remove(correct_conf.Name())
	defer os.Remove(wrong_conf.Name())

	log := logging.MustGetLogger("testing-montrics")
	devnull := logging.NewLogBackend(os.Stderr, "", 1)
	logging.SetBackend(devnull)

	z, err := NewServer("sdsdsd", log)
	if err == nil {
		t.Logf("Created a server type without a valid config file: %v", z)
		t.Fail()
	}
	z, err = NewServer(wrong_conf.Name(), log)
	if err == nil {
		t.Log("Somehow created a server type with the wrong config")
		t.Fail()
	}
	z, err = NewServer(correct_conf.Name(), log)
	if err == nil {
		t.Log("Created a server type without action files")
		t.Fail()
	}

}

func TestParseActions(t *testing.T) {
	empty_server := &Server{actions: []*Action{}, metrics: []*Metric{}}
	correct_file, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = correct_file.WriteString(action_yml)
	wrong_file, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = wrong_file.WriteString("---{broken yml\n:")
	if err != nil {
		t.Fatal("Failed to setup action files for testing")
	}
	defer os.Remove(correct_file.Name())
	defer os.Remove(wrong_file.Name())
	err = parseActions(correct_file.Name(), empty_server)
	if err != nil {
		t.Log("Failed to parse correct YML")
		t.Fail()
	}
	if len(empty_server.actions) != 1 {
		t.Log("Wrong number of actions after parsing a correct YML")
		t.Fail()
	}
	err = parseActions(wrong_file.Name(), empty_server)
	if err == nil {
		t.Log("Didn't fail to parse a broken YML")
		t.Fail()
	}
	if len(empty_server.actions) != 1 {
		t.Log("Wrong number of actions after parsing a wrong YML")
		t.Fail()
	}
}

func TestParseMetrics(t *testing.T) {
	empty_server := &Server{actions: []*Action{}, metrics: []*Metric{}}
	correct_file, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = correct_file.WriteString(action_yml)
	wrong_file, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = wrong_file.WriteString("---{broken yml\n:")
	if err != nil {
		t.Fatal("Failed to setup action files for testing")
	}
	defer os.Remove(correct_file.Name())
	defer os.Remove(wrong_file.Name())

	err = parseMetrics(correct_file.Name(), empty_server)
	if err != nil {
		t.Log("Failed to parse correct YML")
		t.Fail()
	}
	if len(empty_server.metrics) != 1 {
		t.Log("Wrong number of metrics after parsing a correct YML")
		t.Fail()
	}
	err = parseMetrics(wrong_file.Name(), empty_server)
	if err == nil {
		t.Log("Didn't fail to parse a broken YML")
		t.Fail()
	}
	if len(empty_server.metrics) != 1 {
		t.Log("Wrong number of actions after parsing a wrong YML")
		t.Fail()
	}
}

func BenchmarkParseActions(b *testing.B) {
	empty_server := &Server{actions: []*Action{}, metrics: []*Metric{}}
	dummy, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = dummy.WriteString("---\n")
	for i := 0; i < 10000; i++ {
		_, err = dummy.WriteString(fmt.Sprintf("- handler:\n    action: campfire%d\n", i))
		_, err = dummy.WriteString(`    vars:
      rooms: "Operations Room,Ops Tears"
      campfire_api_key: "asdkAJDKJ#K@JK#JDK@J#DKJ@"
      campfire_subdomain: "testing"
`)
	}
	if err != nil {
		b.Log("Cold not create dummy thousand actions")
		b.FailNow()
	}
	defer os.Remove(dummy.Name())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = parseActions(dummy.Name(), empty_server)
	}
	if err != nil {
		b.Log("Failed to parse actions")
		b.FailNow()
	}
	b.Logf("Number of actions: %d", len(empty_server.actions))
}
