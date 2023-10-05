package middlewares

import (
	"fmt"
	"net"
	"net/http"

	"go.uber.org/zap"
)

const (
	ipHeaderName = "X-Real-IP"
)

// CheckSubnet checks IP address.
func checkSubnet(subnet *net.IPNet, r *http.Request) error {
	if subnet == nil {
		return nil
	}
	var ip net.IP
	if r.Header.Get(ipHeaderName) == "" {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return fmt.Errorf("subnet checker ip ('%s') parse error: %w", r.RemoteAddr, err)
		}
		ip = net.ParseIP(host)
	} else {
		ip = net.ParseIP(r.Header.Get(ipHeaderName))
	}
	if !subnet.Contains(ip) {
		return fmt.Errorf("subnet checker error: ip ('%s') request rejected", ip)
	}
	return nil
}

// SubNetCheckMiddleware checks request IP in Header "X-Real-IP".
func SubNetCheckMiddleware(
	subnet *net.IPNet,
	logger *zap.SugaredLogger,
) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if err := checkSubnet(subnet, r); err != nil {
				w.WriteHeader(http.StatusForbidden)
				logger.Warnln(err.Error())
			} else {
				next.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	}
}
