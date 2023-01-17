package zoox

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-zoox/core-utils/object"
	"github.com/spf13/cast"
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
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}

	requestIDPrefix = fmt.Sprintf("%s/%s", hostname, b64[0:10])
}

// ZValue is the value string to cast to new type with spf13/cast.
type ZValue string

// ToInt ...
func (s ZValue) ToInt() int {
	return cast.ToInt(s)
}

// ToInt64 ...
func (s ZValue) ToInt64() int64 {
	return cast.ToInt64(s)
}

// ToUInt ...
func (s ZValue) ToUInt() uint {
	return cast.ToUint(s)
}

// ToUint64 ...
func (s ZValue) ToUint64() uint64 {
	return cast.ToUint64(s)
}

// ToBool ...
func (s ZValue) ToBool() bool {
	return cast.ToBool(s)
}

// ToFloat64 ...
func (s ZValue) ToFloat64() float64 {
	return cast.ToFloat64(s)
}

// ToTime ...
func (s ZValue) ToTime() time.Time {
	return cast.ToTime(s)
}

// ToDuration ...
func (s ZValue) ToDuration() time.Duration {
	return cast.ToDuration(s)
}

// ToString ...
func (s ZValue) ToString() string {
	return string(s)
}

// Query ...
type Query struct {
	ctx *Context
}

func newQuery(ctx *Context) *Query {
	return &Query{
		ctx: ctx,
	}
}

// Get gets request query with the given name.
func (q *Query) Get(key string, defaultValue ...string) ZValue {
	value := q.ctx.Request.URL.Query().Get(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return ZValue(value)
}

// Param ...
type Param struct {
	ctx *Context
	//
	params map[string]string
}

func newParams(ctx *Context, value map[string]string) *Param {
	return &Param{
		ctx:    ctx,
		params: value,
	}
}

// Get gets request param with the given name.
func (q *Param) Get(key string, defaultValue ...string) ZValue {
	value, ok := q.params[key]
	if ok {
		return ZValue(value)
	}

	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return ZValue(value)
}

// Iterator ...
func (q *Param) Iterator() map[string]string {
	return q.params
}

// Form ...
type Form struct {
	ctx *Context
	//
	params map[string]string
}

func newForm(ctx *Context) *Form {
	return &Form{
		ctx:    ctx,
		params: make(map[string]string),
	}
}

// Get gets request form with the given name.
func (f *Form) Get(key string, defaultValue ...string) string {
	value := f.ctx.Request.FormValue(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
}

// Body ...
type Body struct {
	ctx *Context
	//
	data map[string]interface{}
}

func newBody(ctx *Context) *Body {
	return &Body{
		ctx: ctx,
	}
}

// Get gets request form with the given name.
func (f *Body) Get(key string, defaultValue ...interface{}) interface{} {
	if f.data == nil {
		f.data = f.ctx.Bodies()
	}

	value := object.Get(f.data, key)

	// @TODO generic cannot compare zero value
	// if value == "" && len(defaultValue) > 0 {
	// 	value = defaultValue[0]
	// }

	return value
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

	colon := strings.LastIndexByte(host, ':')
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
