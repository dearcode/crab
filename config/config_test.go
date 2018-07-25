package config

import (
	"os"
	"testing"

	"github.com/juju/errors"
)

var (
	c    *Config
	data = `
 [db]
domain    =mailchina.org
db_enable=true
# test comments
;port=3306
[api]
   url=http://baiud.com/
enable=T
headersize=123
`
)

func TestMain(main *testing.M) {
	nc, err := NewConfig("", data)
	if err != nil {
		panic(err.Error())
	}
	c = nc

	os.Exit(main.Run())
}

func TestConfigString(t *testing.T) {
	var domain string
	if err := c.GetData("db", "domain", &domain, "test.db.com"); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("domain:%v", domain)
}

func TestConfigInt(t *testing.T) {
	var port int
	if err := c.GetData("db", "port", &port, 3359); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("port:%d", port)
}

func TestConfigBool(t *testing.T) {
	var enable bool
	if err := c.GetData("db", "enable", &enable, false); err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("enable:%v", enable)
}

type testConf struct {
	DB struct {
		Domain string
		Port   int  `cfg_default:"9088"`
		Enable bool `cfg_key:"db_enable"`
	}

	API struct {
		URL        string
		Enable     bool
		HeaderSize int
	}
}

func TestConfigStruct(t *testing.T) {
	var conf testConf
	if err := ParseConfig(data, &conf); err != nil {
		t.Fatalf(errors.ErrorStack(err))
	}
	t.Logf("conf:%+v", conf)
}
