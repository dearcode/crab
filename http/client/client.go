package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/juju/errors"

	"github.com/dearcode/crab/log"
)

//HTTPClient 带超时重试控制的http客户端.
type HTTPClient struct {
	retryTimes int
	timeout    time.Duration
	client     http.Client
	logger     *log.Logger
}

type StatusError struct {
	Code    int
	Message string
}

func (se *StatusError) Error() string {
	return fmt.Sprintf("HTTP Status %v %s", se.Code, se.Message)
}

func (c *HTTPClient) dial(network, addr string) (net.Conn, error) {
	var conn net.Conn
	var err error

	for i := 0; i < c.retryTimes; i++ {
		conn, err = net.DialTimeout(network, addr, c.timeout)
		if err == nil {
			break
		}
		c.logger.Errorf("DialTimeout %s:%s error:%v retry:%v", network, addr, err, i+1)
	}

	if err != nil {
		c.logger.Errorf("DialTimeout %s:%s error:%v", network, addr, err)
		return nil, errors.Trace(err)
	}

	deadline := time.Now().Add(c.timeout)
	if err = conn.SetDeadline(deadline); err != nil {
		c.logger.Errorf("SetDeadline %s:%s", network, addr)
		conn.Close()
		return nil, errors.Trace(err)
	}

	return conn, nil
}

const (
	defaultRetryTimes = 3
	defaultTimeout    = 300
)

//New 创建一个带超时和重试控制的http client, 单位秒.
func New() *HTTPClient {
	hc := &HTTPClient{
		client:     http.Client{},
		timeout:    time.Duration(defaultTimeout) * time.Second,
		retryTimes: defaultRetryTimes,
	}

	hc.client.Transport = &http.Transport{Dial: hc.dial}
	return hc
}

//Timeout 设置超时时间，单位:秒, 默认300秒.
func (c *HTTPClient) Timeout(t int) *HTTPClient {
	c.timeout = time.Duration(t) * time.Second
	return c
}

//SetLogger 开启日志.
func (c *HTTPClient) SetLogger(l *log.Logger) *HTTPClient {
	c.logger = l
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

	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") || strings.Contains(resp.Header.Get("Content-Type"), "gzip") {
		gr, err := gzip.NewReader(bytes.NewBuffer(data))
		if err != nil {
			return nil, errors.Trace(err)
		}
		defer gr.Close()
		data, err = ioutil.ReadAll(gr)
		if err != nil {
			return nil, errors.Trace(err)
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Trace(&StatusError{Code: resp.StatusCode, Message: string(data)})
	}

	return data, nil
}

//Get 发送Get请求.
func (c HTTPClient) Get(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("GET", url, headers, body)
}

//GetJSON 发送Get请求, 并解析返回json.
func (c HTTPClient) GetJSON(url string, headers map[string]string, resp interface{}) error {
	buf, err := c.do("GET", url, headers, nil)
	if err != nil {
		return errors.Trace(err)
	}
	c.logger.Debugf("url:%v, resp:%s", url, buf)
	return errors.Trace(json.Unmarshal(buf, resp))
}

//Post 发Post请求.
func (c HTTPClient) Post(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("POST", url, headers, body)
}

//PostJSON 发送json结构数据请求，并解析返回结果.
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

	c.logger.Debugf("url:%v, req:%+v, resp:%s", url, data, buf)
	return json.Unmarshal(buf, resp)
}

//Put 发送put请求.
func (c HTTPClient) Put(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("PUT", url, headers, body)
}

//PutJSON 发送json请求解析返回结果.
func (c HTTPClient) PutJSON(url string, headers map[string]string, data interface{}, resp interface{}) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return errors.Trace(err)
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-type"] = "application/json"

	if buf, err = c.do("PUT", url, headers, buf); err != nil {
		return errors.Trace(err)
	}
	c.logger.Debugf("url:%v, resp:%s", url, buf)
	return json.Unmarshal(buf, resp)
}

//Delete 发送delete请求.
func (c HTTPClient) Delete(url string, headers map[string]string, body []byte) ([]byte, error) {
	return c.do("DELETE", url, headers, body)
}

//DeleteJSON 发送JSON格式delete请求, 并解析返回结果.
func (c HTTPClient) DeleteJSON(url string, headers map[string]string, resp interface{}) error {
	buf, err := c.do("DELETE", url, headers, nil)
	if err != nil {
		return errors.Trace(err)
	}
	c.logger.Debugf("url:%v, resp:%s", url, buf)
	return json.Unmarshal(buf, resp)
}
