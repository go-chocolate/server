package netutil

import (
	"net"
	"net/http"
	"strings"
)

var (
	defaultTrustedCIDRs = []*net.IPNet{
		{ // 0.0.0.0/0 (IPv4)
			IP:   net.IP{0x0, 0x0, 0x0, 0x0},
			Mask: net.IPMask{0x0, 0x0, 0x0, 0x0},
		},
		{ // ::/0 (IPv6)
			IP:   net.IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			Mask: net.IPMask{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
		},
	}

	defaultTrustedProxies = []string{"0.0.0.0/0", "::/0"}

	defaultForwardedHeader = []string{"X-Forwarded-For", "X-Real-IP"}
)

type IPValidator struct {
	trustedProxies []string
	trustedCIDRs   []*net.IPNet

	// trustedPlatform if set to a constant of value gin.Platform*, trusts the headers set by
	// that platform, for example to determine the client IP
	trustedPlatform string

	// forwardedByClientIP if enabled, client IP will be parsed from the request's headers that
	// match those stored at `(*gin.Engine).RemoteIPHeaders`. If no IP was
	// fetched, it falls back to the IP obtained from
	// `(*gin.Context).Request.RemoteAddr`.
	forwardedByClientIP bool

	// remoteIPHeaders list of headers used to obtain the client IP when
	// `(*gin.Engine).ForwardedByClientIP` is `true` and
	// `(*gin.Context).Request.RemoteAddr` is matched by at least one of the
	// network origins of list defined by `(*gin.Engine).SetTrustedProxies()`.
	remoteIPHeaders []string
}

type IPValidatorOption func(v *IPValidator)

func WithTrustedProxies(proxies []string) IPValidatorOption {
	return func(v *IPValidator) {
		if val, err := prepareTrustedCIDRs(proxies); err == nil {
			v.trustedProxies = proxies
			v.trustedCIDRs = val
		}
	}
}

func WithTrustedPlatform(trustedPlatform string) IPValidatorOption {
	return func(v *IPValidator) {
		v.trustedPlatform = trustedPlatform
	}
}

func WithForwardedByClientIP(forwardedByClientIP bool) IPValidatorOption {
	return func(v *IPValidator) {
		v.forwardedByClientIP = forwardedByClientIP
	}
}

func WithRemoteIPHeaders(remoteIPHeaders []string) IPValidatorOption {
	return func(v *IPValidator) {
		v.remoteIPHeaders = remoteIPHeaders
	}
}

func NewIPValidator(options ...IPValidatorOption) *IPValidator {
	val := &IPValidator{
		trustedProxies:      defaultTrustedProxies,
		trustedCIDRs:        defaultTrustedCIDRs,
		trustedPlatform:     "",
		forwardedByClientIP: true,
		remoteIPHeaders:     defaultForwardedHeader,
	}
	for _, option := range options {
		option(val)
	}
	return val
}

// ClientIP implements one best effort algorithm to return the real client IP.
// It calls c.RemoteIP() under the hood, to check if the remote IP is a trusted proxy or not.
// If it is it will then try to parse the headers defined in Engine.RemoteIPHeaders (defaulting to [X-Forwarded-For, X-Real-Ip]).
// If the headers are not syntactically valid OR the remote IP does not correspond to a trusted proxy,
// the remote IP (coming from Request.RemoteAddr) is returned.
func (v *IPValidator) ClientIP(r *http.Request) string {
	// Check if we're running on a trusted platform, continue running backwards if error
	if v.trustedPlatform != "" {
		// Developers can define their own header of Trusted Platform or use predefined constants
		if addr := requestHeader(r, v.trustedPlatform); addr != "" {
			return addr
		}
	}

	// It also checks if the remoteIP is a trusted proxy or not.
	// In order to perform this validation, it will see if the IP is contained within at least one of the CIDR blocks
	// defined by Engine.SetTrustedProxies()
	remoteIP := net.ParseIP(RemoteIP(r))
	if remoteIP == nil {
		return ""
	}
	trusted := v.isTrustedProxy(remoteIP)

	if trusted && v.forwardedByClientIP && v.remoteIPHeaders != nil {
		for _, headerName := range v.remoteIPHeaders {
			ip, valid := v.validateHeader(requestHeader(r, headerName))
			if valid {
				return ip
			}
		}
	}
	return remoteIP.String()
}

// isTrustedProxy will check whether the IP address is included in the trusted list according to Engine.trustedCIDRs
func (v *IPValidator) isTrustedProxy(ip net.IP) bool {
	if v.trustedCIDRs == nil {
		return false
	}
	for _, cidr := range v.trustedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// validateHeader will parse X-Forwarded-For header and return the trusted client IP address
func (v *IPValidator) validateHeader(header string) (clientIP string, valid bool) {
	if header == "" {
		return "", false
	}
	items := strings.Split(header, ",")
	for i := len(items) - 1; i >= 0; i-- {
		ipStr := strings.TrimSpace(items[i])
		ip := net.ParseIP(ipStr)
		if ip == nil {
			break
		}

		// X-Forwarded-For is appended by proxy
		// Check IPs in reverse order and stop when find untrusted proxy
		if (i == 0) || (!v.isTrustedProxy(ip)) {
			return ipStr, true
		}
	}
	return "", false
}

func requestHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

func prepareTrustedCIDRs(trustedProxies []string) ([]*net.IPNet, error) {

	cidr := make([]*net.IPNet, 0, len(trustedProxies))
	for _, trustedProxy := range trustedProxies {
		if !strings.Contains(trustedProxy, "/") {
			ip := parseIP(trustedProxy)
			if ip == nil {
				return cidr, &net.ParseError{Type: "IP address", Text: trustedProxy}
			}

			switch len(ip) {
			case net.IPv4len:
				trustedProxy += "/32"
			case net.IPv6len:
				trustedProxy += "/128"
			}
		}
		_, cidrNet, err := net.ParseCIDR(trustedProxy)
		if err != nil {
			return cidr, err
		}
		cidr = append(cidr, cidrNet)
	}
	return cidr, nil
}

// parseIP parse a string representation of an IP and returns a net.IP with the
// minimum byte representation or nil if input is invalid.
func parseIP(ip string) net.IP {
	parsedIP := net.ParseIP(ip)

	if ipv4 := parsedIP.To4(); ipv4 != nil {
		// return ip in a 4-byte representation
		return ipv4
	}

	// return ip in a 16-byte representation or nil
	return parsedIP
}

var defaultIPValidator = NewIPValidator()

// ClientIP 使用默认配置获取客户端真实IP
// RemoteIP 为服务端收到客户端连接时识别到的IP地址，但受限于复杂的网络环境，客户端可能使用代理服务，RemoteIP拿到的只是代理服务器IP
// 通常代理服务会在HTTP Header 中追加客户端IP地址，通过读取该Header获取客户端ip
// 同时，代理服务可以设置任意IP，所以对于不信任的代理服务器，可能获取到伪造IP，使用 NewIPValidator 可自定义解析策略
func ClientIP(r *http.Request) string {
	return defaultIPValidator.ClientIP(r)
}
