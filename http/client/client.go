package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/juju/errors"
	"github.com/zssky/log"
)

//HTTPClient 带超时重试控制的http客户端.
type HTTPClient struct {
	retryTimes int
	timeout    time.Duration
	client     http.Client
}

type StatusError struct {
	Code int
}

func (se *StatusError) Error() string {
	return fmt.Sprintf("HTTP Status %v", se.Code)
}

func (c *HTTPClient) Dial(network, addr string) (net.Conn, error) {
	var conn net.Conn
	var err error

	log.Debugf("c:%p, %#v", c, *c)

	for i := 0; i < c.retryTimes; i++ {
		conn, err = net.DialTimeout(network, addr, c.timeout)
		if err == nil {
			break
		}
		log.Errorf("DialTimeout %s:%s error:%v retry:%v", network, addr, err, i+1)
	}

	if err != nil {
		log.Errorf("DialTimeout %s:%s error:%v", network, addr, err)
		return nil, errors.Trace(err)
	}

	deadline := time.Now().Add(c.timeout)
	if err = conn.SetDeadline(deadline); err != nil {
		log.Errorf("SetDeadline %s:%s", network, addr)
		conn.Close()
		return nil, errors.Trace(err)
	}

	return conn, nil
}

const (
	defaultRetryTimes = 3
	defaultTimeout    = 15
)

//New 创建一个带超时和重试控制的http client, 单位秒.
func New() *HTTPClient {
	hc := &HTTPClient{
		client:     http.Client{},
		timeout:    time.Duration(defaultTimeout) * time.Second,
		retryTimes: defaultRetryTimes,
	}

	hc.client.Transport = &http.Transport{Dial: hc.Dial}
	log.Debugf("init hc:%p", hc)
	return hc
}

//Timeout 设置超时时间，单位:秒, 默认15秒.
func (c *HTTPClient) Timeout(t int) *HTTPClient {
	c.timeout = time.Duration(t) * time.Second
	log.Debugf("timeout c:%p", c)
	return c
}

//RetryTimes 设置连接重试次数，默认为3次
func (c *HTTPClient) RetryTimes(t int) *HTTPClient {
	c.retryTimes = t
	return c
}

func (c HTTPClient) do(method, url string, headers map[string]string, body []byte) ([]byte, error) {
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

//Get 发送Get请求.
func (c HTTPClient) Get(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("GET", url, headers, body)
}

//Post 发Post请求.
func (c HTTPClient) Post(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("POST", url, headers, body)
}

//PostJSON 传json结构数据.
func (c HTTPClient) PostJSON(url string, headers map[string]string, data interface{}, resp interface{}) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return errors.Trace(err)
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-type"] = "application/json"

	if buf, err = c.do("POST", url, headers, buf); err != nil {
		return errors.Trace(err)
	}

	return json.Unmarshal(buf, resp)
}

//Put 发送put请求.
func (c HTTPClient) Put(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("PUT", url, headers, body)
}

//Delete 发送delete请求.
func (c HTTPClient) Delete(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("DELETE", url, headers, body)
}
