package aes

import (
	"github.com/juju/errors"
	"testing"
)

func TestAes(t *testing.T) {
	src := "test"
	key := "1qaz@WSX"

	crypted, err := Encrypt(src, key)
	if err != nil {
		t.Fatalf("AesEncrypt error:%s", errors.ErrorStack(err))
	}
	t.Logf("dest:%v", crypted)
	deSrc, err := Decrypt(crypted, key)
	if err != nil {
		t.Fatalf("AesDecrypt error:%s", err.Error())
	}
	t.Logf("desc:%s", deSrc)
}

func TestDesPanic(t *testing.T) {
	key := "1qaz@WSX"
	crypted := "hX0Y5u-EkggQstomvCXgQw=="
	deSrc, err := Decrypt(crypted, key)
	if err != nil {
		t.Logf("err:%v", err)
	}

	t.Logf("%s", deSrc)

}
