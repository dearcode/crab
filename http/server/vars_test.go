package server

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestVarsForm(t *testing.T) {
	url := "http://www.baidu.com/test/"
	body := ioutil.NopCloser(bytes.NewBufferString("email=test@jd.com&passwd=1qaz"))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	result := struct {
		Email    string `json:"email" valid:"Required"`
		Password string `json:"passwd" valid:"Required"`
	}{}

	if err = ParseFormVars(req, &result); err != nil {
		t.Fatal(err.Error())
	}
}

func TestVarsHeader(t *testing.T) {
	url := "http://www.baidu.com/test/"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	req.Header.Add("Email", "test@jd.com")
	req.Header.Add("Passwd", "1qaz")

	result := struct {
		Email    string `json:"Email" valid:"Required"`
		Password string `json:"Passwd" valid:"Required"`
	}{}

	if err = ParseHeaderVars(req, &result); err != nil {
		t.Fatal(err.Error())
	}
}

func TestVarsUrl(t *testing.T) {
	url := "http://www.baidu.com/test/?email=test@jd.com&passwd=1qaz"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err.Error())
	}

	result := struct {
		Email    string `json:"email" valid:"Required"`
		Password string `json:"passwd" valid:"Required"`
	}{}

	if err = ParseURLVars(req, &result); err != nil {
		t.Fatal(err.Error())
	}
}

func TestVarsJSON(t *testing.T) {
	url := "http://www.baidu.com/test/"
	body := ioutil.NopCloser(bytes.NewBufferString(`{"email":"test@jd.com","passwd":"1qaz"}`))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		t.Fatal(err.Error())
	}
	result := struct {
		Email    string `json:"email" valid:"Required"`
		Password string `json:"passwd" valid:"Required"`
	}{}

	if err = ParseJSONVars(req, &result); err != nil {
		t.Fatal(err.Error())
	}
}
