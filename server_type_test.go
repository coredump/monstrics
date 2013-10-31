package monstrics

import (
	"fmt"
	"github.com/op/go-logging"
	"io/ioutil"
	"os"
	"testing"
)

var (
	conf_yml = `---
- action: campfire
  rooms:
    - Operations Room
    - Ops Tears
  api_key: "asdkAJDKJ#K@JK#JDK@J#DKJ@"
  subdomain: "testing"

- path: stats.production.*.unicorn.socket_queued
  constraints:
    Upper Limit: 5
    Lower Limit: 5
`
	main_yml = `---
amqp:
  url: amqp://guest:guest@localhost/
  exchange: metrics
conf_dir: /xxxxx
debug: true
`
	file_prefix = "go_test_monstrics"
	log         = logging.MustGetLogger("testing-montrics")
)

func TestNewServer(t *testing.T) {
	correct_conf, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = correct_conf.WriteString(main_yml)
	wrong_conf, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = wrong_conf.WriteString("---\n{bogus\n:yml")
	if err != nil {
		t.Fatal("Failed to setup conf file")
	}
	defer os.Remove(correct_conf.Name())
	defer os.Remove(wrong_conf.Name())

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
		t.Log("Created a server type without conf files")
		t.Fail()
	}

}

func TestParseActions(t *testing.T) {
	empty_server := &Server{Actions: []*Action{}, Metrics: []*Metric{}}
	correct_file, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = correct_file.WriteString(conf_yml)
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
	if len(empty_server.Actions) != 1 {
		t.Log("Wrong number of Actions after parsing a correct YML")
		t.Fail()
	}
	err = parseActions(wrong_file.Name(), empty_server)
	if err == nil {
		t.Log("Didn't fail to parse a broken YML")
		t.Fail()
	}
	if len(empty_server.Actions) != 1 {
		t.Log("Wrong number of Actions after parsing a wrong YML")
		t.Fail()
	}
}

func TestParseMetrics(t *testing.T) {
	empty_server := &Server{Actions: []*Action{}, Metrics: []*Metric{}}
	correct_file, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = correct_file.WriteString(conf_yml)
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
	if len(empty_server.Metrics) != 1 {
		t.Log("Wrong number of Metrics after parsing a correct YML")
		t.Fail()
	}
	err = parseMetrics(wrong_file.Name(), empty_server)
	if err == nil {
		t.Log("Didn't fail to parse a broken YML")
		t.Fail()
	}
	if len(empty_server.Metrics) != 1 {
		t.Log("Wrong number of Actions after parsing a wrong YML")
		t.Fail()
	}
}

func BenchmarkNewServer(b *testing.B) {
	devnull, _ := os.Open(os.DevNull)
	backend := logging.NewLogBackend(devnull, "", 1)
	logging.SetBackend(backend)

	dummy_dir, err := ioutil.TempDir(os.TempDir(), file_prefix)
	dummy, err := os.Create(fmt.Sprintf("%s/dummy.yml", dummy_dir))
	_, err = dummy.WriteString("---\n")
	for i := 0; i < 10000; i++ {
		_, err = dummy.WriteString(fmt.Sprintf("- action: campfire%d\n", i))
		_, err = dummy.WriteString(`  rooms:
    - Operations Room
    - Ops Tears
  api_key: "asdkAJDKJ#K@JK#JDK@J#DKJ@"
  subdomain: "testing"`)
		_, err = dummy.WriteString(fmt.Sprint("\n"))
		_, err = dummy.WriteString(fmt.Sprintf("- path: stats.production.server%d\n", i))
		_, err = dummy.WriteString(`  constraints:
    Upper Limit: 5
    Lower Limit: 5`)
		_, err = dummy.WriteString(fmt.Sprint("\n"))
	}
	if err != nil {
		b.Log("Cold not create dummy thousand Actions")
		b.FailNow()
	}
	conf_file, err := ioutil.TempFile(os.TempDir(), file_prefix)
	_, err = conf_file.WriteString(fmt.Sprintf("amqp:\n  url: xxx\nconf_dir: %s\n", dummy_dir))
	defer os.RemoveAll(dummy_dir)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server, _ := NewServer(conf_file.Name(), log)
		b.Logf("Numbers: %d, %d", len(server.Actions), len(server.Metrics))
	}
	if err != nil {
		b.Logf("Failed to create server: %v", err)
		b.FailNow()
	}
}
