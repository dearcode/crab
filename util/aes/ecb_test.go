package aes

import (
	"testing"
)

func TestAes(t *testing.T) {
	src := "test"
	key := []byte("dbsjdcom\x00\x00\x00\x00\x00\x00\x00\x00")

	crypted, err := Encrypt(src, key)
	if err != nil {
		t.Fatalf("AesEncrypt error:%s", err.Error())
	}
	deSrc, err := Decrypt(crypted, []byte(key))
	if err != nil {
		t.Fatalf("AesDecrypt error:%s", err.Error())
	}
	t.Logf("desc:%v", deSrc)
}

func TestDesPanic(t *testing.T) {
	key := []byte("dbsjdcom\x00\x00\x00\x00\x00\x00\x00\x00")
	crypted := "fDmxIdK9p3oEyQoL1Bwz4Fakia3Y4Qn1SF8podapMFU="
	deSrc, err := Decrypt(crypted, []byte(key))
	if err != nil {
		t.Logf("err:%v", err)
	}

	t.Logf("%s", deSrc)

}
