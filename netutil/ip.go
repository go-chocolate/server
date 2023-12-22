package netutil

import (
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	envPodIp = "POD_IP"
)

// InternalIp returns an internal ip.
func InternalIp() string {
	infs, err := net.Interfaces()
	if err != nil {
		return ""
	}

	for _, inf := range infs {
		if isEthDown(inf.Flags) || isLoopback(inf.Flags) {
			continue
		}

		addrs, err := inf.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}

	return ""
}

func isEthDown(f net.Flags) bool {
	return f&net.FlagUp != net.FlagUp
}

func isLoopback(f net.Flags) bool {
	return f&net.FlagLoopback == net.FlagLoopback
}

// RemoteIP parses the IP from Request.RemoteAddr, normalizes and returns the IP (without the port).
func RemoteIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		return ""
	}
	return ip
}

// IsPrivate reports whether ip is a private address, according to
// RFC 1918 (IPv4 addresses) and RFC 4193 (IPv6 addresses).
func IsPrivate(ip string) (bool, error) {
	val := net.ParseIP(ip)
	if val != nil {
		return val.IsPrivate(), nil
	}
	return false, errors.New("invalid ip address")
}

func FigureOutListenOn(listenOn string) string {
	addr, err := net.ResolveTCPAddr("", listenOn)
	if err != nil {
		return listenOn
	}
	if ip := addr.IP; len(ip) > 0 && !ip.IsUnspecified() && !ip.IsLoopback() {
		return listenOn
	}

	ip := os.Getenv(envPodIp)
	if len(ip) == 0 {
		ip = InternalIp()
	}
	if len(ip) == 0 {
		return listenOn
	}
	return net.JoinHostPort(ip, strconv.Itoa(addr.Port))
}
