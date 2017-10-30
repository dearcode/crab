package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/juju/errors"
	"github.com/zssky/log"
)

type httpClient struct {
	client http.Client
}

type StatusError struct {
	Code int
}

func (se *StatusError) Error() string {
	return fmt.Sprintf("HTTP Status %v", se.Code)
}

//NewClient 创建一个带超时控制的http client, 单位秒.
func New(ts int) httpClient {
	timeout := time.Duration(ts) * time.Second
	return httpClient{
		client: http.Client{
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

func (c httpClient) do(method, url string, headers map[string]string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, errors.Trace(err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Trace(&StatusError{resp.StatusCode})
	}

	return data, nil
}

func (c httpClient) Get(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("GET", url, headers, body)
}

func (c httpClient) POST(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("POST", url, headers, body)
}

func (c httpClient) PUT(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("PUT", url, headers, body)
}

func (c httpClient) DELETE(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("DELETE", url, headers, body)
}
