package zoox

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync/atomic"

	"github.com/go-zoox/core-utils/object"
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
func (q *Query) Get(key string, defaultValue ...string) string {
	value := q.ctx.Request.URL.Query().Get(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
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
func (q *Param) Get(key string, defaultValue ...string) string {
	value, ok := q.params[key]
	if ok {
		return value
	}

	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
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
