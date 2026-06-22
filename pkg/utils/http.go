package utils

import (
	"net"
	"net/http"
	"strings"
)

// nginx/openresty
// proxy_set_header X-Real-IP $remote_addr;
// proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;

var trustedProxies = mustParseCIDRs(
	"127.0.0.0/8",
	"::1/128",
)

func mustParseCIDRs(cidrs ...string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		_, ipNet, err := net.ParseCIDR(c)
		if err != nil {
			panic("invalid trusted proxy cidr: " + c)
		}
		nets = append(nets, ipNet)
	}
	return nets
}

func isTrustedProxy(ip net.IP) bool {
	for _, n := range trustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func ClientIP(r *http.Request) string {
	remoteIP := remoteAddrIP(r.RemoteAddr)

	if remoteIP == nil || !isTrustedProxy(remoteIP) {
		if remoteIP != nil {
			return remoteIP.String()
		}
		return r.RemoteAddr
	}

	if ip := clientIPFromXFF(r.Header.Get("X-Forwarded-For")); ip != "" {
		return ip
	}

	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		if ip := net.ParseIP(xrip); ip != nil {
			return ip.String()
		}
	}

	return remoteIP.String()
}

func clientIPFromXFF(xff string) string {
	if xff == "" {
		return ""
	}
	parts := strings.Split(xff, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(parts[i])
		ip := net.ParseIP(candidate)
		if ip == nil {
			continue
		}
		if !isTrustedProxy(ip) {
			return ip.String()
		}
	}
	return ""
}
func remoteAddrIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return net.ParseIP(host)
}

func UserAgent(r *http.Request) string {
	return strings.TrimSpace(r.UserAgent())
}
