package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/juju/errors"
)

var (
	c    *Config
	path string
)

func TestMain(main *testing.M) {
	data := `
 [db]
domain    =jd.com
enable=true
# test comments
;port=3306
[api]
   url=http://baiud.com/
enable=T
headersize=123
`
	f, err := ioutil.TempFile("/tmp/", "test_config_")
	if err != nil {
		panic(err.Error())
	}
	f.WriteString(data)
	path = f.Name()
	f.Close()

	if c, err = NewConfig(path); err != nil {
		panic(err.Error())
	}

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
		Port   int `default:"9088"`
		Enable bool
	}

	API struct {
		URL        string
		Enable     bool
		HeaderSize int
	}

	aaa int
}

func TestConfigStruct(t *testing.T) {
	var conf testConf
	if err := LoadConfig(path, &conf); err != nil {
		t.Fatalf(errors.ErrorStack(err))
	}
	t.Logf("conf:%+v", conf)
}
