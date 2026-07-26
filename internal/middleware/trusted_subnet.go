package middleware

import (
	"net"
	"net/http"
)

// RealIPHeader is the request header carrying the client IP address that the
// TrustedSubnet middleware checks against the trusted subnet.
const RealIPHeader = "X-Real-IP"

// TrustedSubnet returns middleware that admits a request only when the client
// IP taken from the X-Real-IP header falls inside the given CIDR subnet. An
// empty or unparseable cidr denies every request, so the endpoint stays closed
// unless a valid trusted subnet is configured. Rejected requests receive
// 403 Forbidden.
func TrustedSubnet(cidr string) func(http.Handler) http.Handler {
	var subnet *net.IPNet
	if cidr != "" {
		if _, parsed, err := net.ParseCIDR(cidr); err == nil {
			subnet = parsed
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if subnet == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(r.Header.Get(RealIPHeader))
			if ip == nil || !subnet.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
