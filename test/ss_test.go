package test

import (
	"github.com/hlhgogo/xsocks-go/adapater/shadowsocks"
	"log"
	"testing"
)

func TestSS(t *testing.T) {
	method := "CHACHA20-IETF-POLY1305"
	port := 37346
	password := "transocks"
	ss := shadowsocks.NewSS(port, method, password)
	if err := ss.RunTcp(); err != nil {
		t.Error(err)
		return
	}

	if err := ss.RunUDP(); err != nil {
		t.Error(err)
		return
	}

	log.Printf("Start shadowsocks [%s|%s|%d]", method, password, port)
	select {}
}
