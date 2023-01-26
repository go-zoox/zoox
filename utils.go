package zoox

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	gostrings "strings"
	"sync/atomic"

	"github.com/go-zoox/core-utils/strings"
)

// RequestIDHeader is the name of the header that contains the request ID.
var RequestIDHeader = "X-Request-Id"

var requestIDPrefix string
var requestID int64
var hostname string

func init() {
	hostname, err := os.Hostname()
	if hostname == "" || err != nil {
		hostname = "localhost"
	}

	var buf [12]byte
	var b64 string
	if len(b64) < 10 {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = gostrings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	requestIDPrefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}

// GenerateRequestID generates a unique request ID.
func GenerateRequestID() string {
	myID := atomic.AddInt64(&requestID, 1)
	return fmt.Sprintf("%s-%06d", requestIDPrefix, myID)
}

// splitHostPort separates host and port. If the port is not valid, it returns
// the entire input as host, and it doesn't check the validity of the host.
// Unlike net.SplitHostPort, but per RFC 3986, it requires ports to be numeric.
func splitHostPort(hostPort string) (host, port string) {
	host = hostPort

	colon := gostrings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}
