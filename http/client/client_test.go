package client

import (
	"testing"
)

func TestHTTPClient(t *testing.T) {
	hc := New().Timeout(1)
	buf, err := hc.Get("http://baidu.com/", nil, nil)
	if err != nil {
		t.Fatalf("err:%v", err)
	}

	t.Logf("buf:%s", buf)

}
