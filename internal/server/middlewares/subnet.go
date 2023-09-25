package middlewares

import (
	"net"
	"net/http"

	"go.uber.org/zap"
)

const (
	ipHeaderName = "X-Real-IP"
)

// SubNetCheckMiddleware checks request IP in Header "X-Real-IP".
func SubNetCheckMiddleware(
	subnet *net.IPNet,
	logger *zap.SugaredLogger,
) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if subnet != nil {
				var ip net.IP
				if r.Header.Get(ipHeaderName) == "" {
					host, _, err := net.SplitHostPort(r.RemoteAddr)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						logger.Warnf("subnet checker ip ('%s') parse error: %v", r.RemoteAddr, err)
						return
					}
					ip = net.ParseIP(host)
				} else {
					ip = net.ParseIP(r.Header.Get(ipHeaderName))
				}
				if !subnet.Contains(ip) {
					w.WriteHeader(http.StatusForbidden)
					logger.Infof("subnet checker error: ip ('%s') request rejected", ip)
					return
				}
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
