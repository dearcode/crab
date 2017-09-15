package httpClient

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/juju/errors"
	"github.com/zssky/log"
)

//Client 对http client简单封装.
type Client struct {
	hc http.Client
}

//NewClient 创建一个带超时控制的http client.
func NewClient(timeout time.Duration) Client {
	return Client{
		hc: http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, timeout)
					if err != nil {
						log.Errorf("DialTimeout %s:%s", netw, addr)
						return nil, errors.Trace(err)
					}
					deadline := time.Now().Add(timeout)
					if err = c.SetDeadline(deadline); err != nil {
						log.Errorf("SetDeadline %s:%s", netw, addr)
						return nil, errors.Trace(err)
					}
					return c, nil
				},
			},
		},
	}
}

func (c Client) do(method, url string, headers map[string]string, body *bytes.Buffer) ([]byte, int, error) {
	var req *http.Request
	var err error

	//参数body就个指向结构体的指针(*bytes.Buffer)，NewRequest的body参数是一个接口(io.Reader)
	if body == nil {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, body)
	}

	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, 0, errors.Trace(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, errors.Trace(err)
	}

	return data, resp.StatusCode, nil
}

//Get get 请求...
func (c Client) Get(url string, headers map[string]string, body *bytes.Buffer) ([]byte, int, error) {
	return c.do("GET", url, headers, body)
}

//POST post 请求.
func (c Client) POST(url string, headers map[string]string, body *bytes.Buffer) ([]byte, int, error) {
	return c.do("POST", url, headers, body)
}

//PUT put 请求.
func (c Client) PUT(url string, headers map[string]string, body *bytes.Buffer) ([]byte, int, error) {
	return c.do("PUT", url, headers, body)
}

//DELETE delete 请求.
func (c Client) DELETE(url string, headers map[string]string, body *bytes.Buffer) ([]byte, int, error) {
	return c.do("DELETE", url, headers, body)
}
