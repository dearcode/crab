package client

import (
	"dearcode.net/crab/log"
	"testing"
)

func TestHTTPClientZIP(t *testing.T) {
	hc := New().SetLogger(log.GetLogger()).Timeout(1)
	buf, err := hc.Get("http://www.baidu.com/", map[string]string{"Accept-Encoding": "gzip"}, nil)
	if err != nil {
		t.Fatalf("err:%v", err)
	}

	t.Logf("buf:%s", buf)
}

func TestHTTPClient(t *testing.T) {
	hc := New().SetLogger(log.GetLogger()).Timeout(1)
	buf, err := hc.Get("http://www.baidu.com/", nil, nil)
	if err != nil {
		t.Fatalf("err:%v", err)
	}

	t.Logf("buf:%s", buf)

}
