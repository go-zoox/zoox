package zoox

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

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

type Query struct {
	ctx *Context
}

func newQuery(ctx *Context) *Query {
	return &Query{
		ctx: ctx,
	}
}

func (q *Query) Get(key string, defaultValue ...string) string {
	value := q.ctx.Request.URL.Query().Get(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
}

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

func (q *Param) Iterator() map[string]string {
	return q.params
}

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

func (f *Form) Get(key string, defaultValue ...string) string {
	value := f.ctx.Request.FormValue(key)
	if value == "" && len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	return value
}

func GenerateRequestID() string {
	myID := atomic.AddInt64(&requestID, 1)
	return fmt.Sprintf("%s-%06d", requestIDPrefix, myID)
}
