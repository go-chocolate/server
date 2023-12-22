package netutil

import (
	"net/http"
	"testing"
)

func build(remote string, header ...string) *http.Request {
	r := new(http.Request)
	r.RemoteAddr = remote
	r.Header = http.Header{}
	if len(header)%2 == 0 {
		for i := 0; i < len(header); i += 2 {
			r.Header.Set(header[i], header[i+1])
		}
	}
	return r
}

func TestClientIP(t *testing.T) {
	r1 := build("123.5.12.3:7523")
	r2 := build("123.5.12.3:7523", "X-Forwarded-For", "213.20.35.100,23.37.125.200")
	r3 := build("123.5.12.3:7523", "X-Real-IP", "178.98.10.2")
	r4 := build("123.5.12.3:7523", "X-Forwarded-For", "23.37.125.200", "X-Real-IP", "178.98.10.2")

	if ClientIP(r1) != "123.5.12.3" {
		t.Fail()
	}
	if ClientIP(r2) != "213.20.35.100" {
		t.Fail()
	}
	if ClientIP(r3) != "178.98.10.2" {
		t.Fail()
	}
	if ClientIP(r4) != "23.37.125.200" {
		t.Fail()
	}
}

func TestValidator(t *testing.T) {
}
