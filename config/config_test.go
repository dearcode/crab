package config

import (
	"os"
	"testing"
)

var (
	c *Config
)

func TestMain(main *testing.M) {
	var err error
	if c, err = NewConfig("./test.ini"); err != nil {
		panic(err.Error())
	}
	os.Exit(main.Run())
}

func TestConfigString(t *testing.T) {
	var domain string
	if err := c.GetData("db", "domain", &domain); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("domain:%v", domain)
}

func TestConfigInt(t *testing.T) {
	var port int
	if err := c.GetData("db", "port", &port); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("port:%d", port)
}

func TestConfigBool(t *testing.T) {
	var enable bool
	if err := c.GetData("db", "enable", &enable); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("enable:%v", enable)
}

type testConf struct {
	DB struct {
		Domain string
		Port   int
		Enable bool
	}

	API struct {
		URL        string
		Enable     bool
		HeaderSize int
	}
}

func TestConfigStruct(t *testing.T) {
	var conf testConf
	if err := LoadConfig("./test.ini", &conf); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("conf:%+v", conf)
}
