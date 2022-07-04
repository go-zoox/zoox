package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/go-zoox/zoox"
)

var trueClientIP = http.CanonicalHeaderKey("True-Client-IP")
var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

func RealIP() zoox.Middleware {
	return func(ctx *zoox.Context) {
		if rip := realIP(ctx.Request); rip != "" {
			ctx.Request.RemoteAddr = rip
		}

		ctx.Next()
	}
}

func realIP(r *http.Request) string {
	var ip string

	if tcip := r.Header.Get(trueClientIP); tcip != "" {
		ip = tcip
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ",")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	}

	if ip == "" || net.ParseIP(ip) == nil {
		return ""
	}

	return ip
}
