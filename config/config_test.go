package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	c, err := NewConfig("./test.ini")
	if err != nil {
		t.Fatalf(err.Error())
	}

    t.Logf("config:%+v", *c)
    var domain string
    if err = c.GetData("db", "domain", &domain); err != nil {
		t.Fatalf(err.Error())
    }
    t.Logf("domain:%v", domain)

}
