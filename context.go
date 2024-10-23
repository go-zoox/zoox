package zoox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/go-errors/errors"

	rd "runtime/debug"

	"time"

	"github.com/go-zoox/cache"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/i18n"
	"github.com/go-zoox/proxy"
	"github.com/go-zoox/zoox/components/application/cmd"
	"github.com/go-zoox/zoox/components/application/cron"
	"github.com/go-zoox/zoox/components/application/debug"
	"github.com/go-zoox/zoox/components/application/env"
	"github.com/go-zoox/zoox/components/application/jobqueue"
	"github.com/go-zoox/zoox/components/context/body"
	"github.com/go-zoox/zoox/components/context/form"
	"github.com/go-zoox/zoox/components/context/mq"
	"github.com/go-zoox/zoox/components/context/param"
	"github.com/go-zoox/zoox/components/context/pubsub"
	"github.com/go-zoox/zoox/components/context/query"
	"github.com/go-zoox/zoox/components/context/sse"
	"github.com/go-zoox/zoox/components/context/state"
	"github.com/go-zoox/zoox/components/context/user"
	"github.com/go-zoox/zoox/utils"

	"github.com/go-zoox/concurrency"
	"github.com/go-zoox/cookie"
	"github.com/go-zoox/core-utils/safe"
	"github.com/go-zoox/core-utils/strings"
	"github.com/go-zoox/fetch"
	"github.com/go-zoox/headers"
	"github.com/go-zoox/jwt"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/random"
	"github.com/go-zoox/session"
	"github.com/go-zoox/tag"
	"github.com/go-zoox/tag/datasource"
	"gopkg.in/yaml.v3"
)

// Context is the request context
type Context struct {
	// Writer is the response writer.
	Writer ResponseWriter

	// Request is the original request object.
	Request *http.Request
	// Request is the alias of Writer.
	Response ResponseWriter

	// request
	Method string
	Path   string
	//
	param param.Param

	query query.Query

	form form.Form

	body body.Body

	// response
	sse sse.SSE

	//
	cookie  cookie.Cookie
	session session.Session
	jwt     jwt.Jwt
	//
	cache cache.Cache
	cron  cron.Cron
	queue jobqueue.JobQueue
	//
	i18n i18n.I18n
	//
	pubsub pubsub.PubSub
	mq     mq.MQ
	//
	env   env.Env
	debug debug.Debug
	// middleware
	handlers []HandlerFunc
	index    int
	//
	App *Application
	//
	Logger *logger.Logger
	//
	//
	state state.State
	user  user.User
	//
	cmd cmd.Cmd
	// request id
	requestID string

	//
	isUpgrade    bool
	isUpgradeSet bool

	// bodyBytes is used to copy body
	bodyBytes []byte

	// once
	once struct {
		debug sync.Once
		//
		cache sync.Once
		queue sync.Once
		env   sync.Once
		//
		i18n sync.Once
		//
		pubsub sync.Once
		mq     sync.Once
		//
		cron sync.Once
		jwt  sync.Once
		sse  sync.Once
		//
		cookie  sync.Once
		session sync.Once
		//
		query sync.Once
		form  sync.Once
		body  sync.Once
		//
		state sync.Once
		user  sync.Once
		//
		cmd sync.Once
	}
}

func newContext(app *Application, w http.ResponseWriter, req *http.Request) *Context {
	// path := strings.TrimSuffix(req.URL.Path, "/")
	path := req.URL.Path
	// path := req.URL.Path
	// if !strings.HasSuffix(path, "/") {
	// 	path += "/"
	// }

	writer := newResponseWriter(w)
	ctx := &Context{
		App: app,
		//
		Writer: writer,
		//
		Request:  req,
		Response: writer,
		//
		Method: req.Method,
		Path:   path,
		//
		index: -1,
	}
	//

	ctx.requestID = ctx.Header().Get(utils.RequestIDHeader)
	if ctx.requestID == "" {
		ctx.requestID = utils.GenerateRequestID()
	}

	ctx.Logger = logger.New(func(opt *logger.Option) {
		// fmt.Println("ctx.Logger:", app.Config.LogLevel)
		opt.Level = app.Config.LogLevel
	})

	return ctx
}

// Context returns the context
func (ctx *Context) Context() context.Context {
	return ctx.Request.Context()
}

// Next runs the next handler in the middleware stack
func (ctx *Context) Next() {
	ctx.index++
	s := len(ctx.handlers)
	// for ; ctx.index < s; ctx.index ++ {
	// 	ctx.handlers[ctx.index](ctx)
	// }

	if ctx.index >= s {
		panic("Handler cannot call ctx.Next")
	}

	ctx.handlers[ctx.index](ctx)
}

// Query returns the query string parameter with the given name.
func (ctx *Context) Query() query.Query {
	ctx.once.query.Do(func() {
		ctx.query = query.New(ctx.Request)
	})

	return ctx.query
}

// Param returns the named URL parameter value if it exists.
func (ctx *Context) Param() param.Param {
	return ctx.param
}

// Header gets the header value by key.
func (ctx *Context) Header() http.Header {
	return ctx.Request.Header
}

// Form returns the form data from POST or PUT request body.
func (ctx *Context) Form() form.Form {
	ctx.once.form.Do(func() {
		ctx.form = form.New(ctx.Request)
	})

	return ctx.form
}

// Body returns the request body.
func (ctx *Context) Body() body.Body {
	ctx.once.body.Do(func() {
		ctx.body = body.New(ctx.Bodies)
	})

	return ctx.body
}

// Status sets the HTTP response status code.
func (ctx *Context) Status(status int) {
	ctx.Writer.WriteHeader(status)
}

// StatusCode returns the HTTP response status code.
func (ctx *Context) StatusCode() int {
	return ctx.Writer.Status()
}

// Get alias for ctx.Header.
func (ctx *Context) Get(key string) string {
	return ctx.Header().Get(key)
}

// Set alias for ctx.SetHeader.
func (ctx *Context) Set(key string, value string) {
	ctx.SetHeader(key, value)
}

// SetHeader sets a header in the response.
func (ctx *Context) SetHeader(key string, value string) {
	ctx.Writer.Header().Set(key, value)
}

// AddHeader adds a header to the response.
func (ctx *Context) AddHeader(key string, value string) {
	ctx.Writer.Header().Add(key, value)
}

// SSE sets the response header for server-sent events.
func (ctx *Context) SSE() sse.SSE {
	ctx.once.sse.Do(func() {
		ctx.sse = sse.New(ctx.Writer)
	})

	return ctx.sse
}

// BasicAuth returns the user/password pair for Basic Authentication.
func (ctx *Context) BasicAuth() (username string, password string, ok bool) {
	return ctx.Request.BasicAuth()
}

// Authorization returns the authorization header for auth.
func (ctx *Context) Authorization() string {
	return ctx.Header().Get(headers.Authorization)
}

// BearerToken returns the token for bearer authentication.
func (ctx *Context) BearerToken() (token string, ok bool) {
	authorization := ctx.Authorization()
	if len(authorization) < 8 {
		return "", false
	}

	return authorization[7:], true
}

// Accept returns the request accept header.
func (ctx *Context) Accept() string {
	return ctx.Header().Get(headers.Accept)
}

// AcceptLanguage returns the request accept header.
func (ctx *Context) AcceptLanguage() string {
	return ctx.Header().Get(headers.AcceptLanguage)
}

// AcceptEncoding returns the request accept header.
func (ctx *Context) AcceptEncoding() string {
	return ctx.Header().Get(headers.AcceptEncoding)
}

// Connection return the request connection header.
func (ctx *Context) Connection() string {
	return ctx.Header().Get(headers.Connection)
}

// UserAgent return the request user-agent header.
func (ctx *Context) UserAgent() string {
	return ctx.Header().Get(headers.UserAgent)
}

// ContentType return the request content-type header.
func (ctx *Context) ContentType() string {
	return ctx.Header().Get(headers.ContentType)
}

// XForwardedFor return the request x-forwarded-for header.
func (ctx *Context) XForwardedFor() string {
	return ctx.Header().Get(headers.XForwardedFor)
}

// XForwardedProto return the request x-forwarded-proto header.
func (ctx *Context) XForwardedProto() string {
	return ctx.Header().Get(headers.XForwardedProto)
}

// XForwardedHost return the request x-forwarded-host header.
func (ctx *Context) XForwardedHost() string {
	return ctx.Header().Get(headers.XForwardedHost)
}

// XForwardedPort return the request x-forwarded-port header.
func (ctx *Context) XForwardedPort() string {
	return ctx.Header().Get(headers.XForwardedPort)
}

// XRealIP return the request x-real-ip header.
func (ctx *Context) XRealIP() string {
	return ctx.Header().Get(headers.XRealIP)
}

// Upgrade return the request upgrade header.
func (ctx *Context) Upgrade() string {
	return ctx.Header().Get(headers.Upgrade)
}

// Origin returns the origin of the request.
func (ctx *Context) Origin() string {
	return ctx.Header().Get(headers.Origin)
}

// Referrer returns the referrer of the request.
func (ctx *Context) Referrer() string {
	return ctx.Header().Get(headers.Referrer)
}

// IsConnectionUpgrade checks if the connection upgrade.
func (ctx *Context) IsConnectionUpgrade() bool {
	if !ctx.isUpgradeSet {
		ctx.isUpgradeSet = true
		ctx.isUpgrade = strings.ToLower(ctx.Connection()) == "upgrade"
	}

	return ctx.isUpgrade
}

// Write writes the data to the connection.
func (ctx *Context) Write(b []byte) {
	ctx.Writer.Write(b)
}

// String writes the given string to the response.
func (ctx *Context) String(status int, format string, values ...interface{}) {
	ctx.Status(status)
	ctx.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON serializes the given struct as JSON into the response body.
func (ctx *Context) JSON(status int, obj interface{}) {
	ctx.Status(status)
	ctx.SetHeader(headers.ContentType, "application/json")
	encoder := json.NewEncoder(ctx.Writer)
	if err := encoder.Encode(obj); err != nil {
		// ctx.Error(http.StatusInternalServerError, err.Error())

		ctx.Logger.Errorf("[ctx.JSON] encode error: %s", err)
		ctx.String(http.StatusInternalServerError, err.Error())
	}
}

// Data writes some data into the body stream and updates the HTTP code.
// Align to gin framework.
func (ctx *Context) Data(status int, contentType string, data []byte) {
	ctx.Status(status)
	ctx.SetHeader(headers.ContentType, contentType)
	ctx.Write(data)
}

// HTML renders the given template with the given data and writes the result
func (ctx *Context) HTML(status int, html string, data ...any) {
	ctx.Template(status, func(tc *TemplateConfig) {
		tc.ContentType = "text/html"
		tc.Content = html
		if len(data) > 0 {
			tc.Data = data[0]
		}
	})
}

// Render renders a template with data and writes the result to the response.
func (ctx *Context) Render(status int, name string, data interface{}) {
	ctx.Template(status, func(tc *TemplateConfig) {
		tc.Name = name
		tc.Data = data
	})
}

// RenderHTML renders a template with data and writes the result to the response.
func (ctx *Context) RenderHTML(filepath string) {
	if !strings.StartsWith(filepath, "/") {
		filepath = fs.JoinCurrentDir(filepath)
	}

	cacheKey := fmt.Sprintf("static_fs:%s", filepath)
	html := ""
	err := ctx.Cache().Get(cacheKey, &html)
	if err != nil {
		html, err = fs.ReadFileAsString(filepath)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, fmt.Errorf("failed to read index.html: %s", err).Error())
			return
		}

		ctx.Cache().Set(cacheKey, &html, 60*time.Second)
	}

	ctx.HTML(http.StatusOK, html)
}

// RenderIndexHTML renders the index.html file from the static directory.
func (ctx *Context) RenderIndexHTML(dir string) {
	ctx.RenderHTML(fs.JoinPath(dir, "index.html"))
}

// RenderStatic renders the static file from the static directory.
func (ctx *Context) RenderStatic(prefix, dir string) {
	hf := http.Dir(dir)
	hfs := http.StripPrefix(prefix, http.FileServer(hf))

	hfs.ServeHTTP(ctx.Writer, ctx.Request)
}

// Error writes the given error to the response.
// Use for system errors
//  1. Internal server error
//  2. Not found
func (ctx *Context) Error(status int, message string) {
	// ctx.Status(status)
	// ctx.Write([]byte(message))

	if ctx.AcceptJSON() {
		ctx.JSON(status, H{
			"code":      status,
			"message":   message,
			"method":    ctx.Method,
			"path":      ctx.Path,
			"timestamp": time.Now(),
		})
		return
	}

	ctx.String(status, message)
}

// Success writes the given data with code-message-result specification to the response.
func (ctx *Context) Success(result interface{}) {
	ctx.JSON(http.StatusOK, H{
		"code":    200,
		"message": "success",
		"result":  result,
	})
}

// Fail writes the given error with code-message-result specification to the response.
func (ctx *Context) Fail(err error, code int, message string, status ...int) {
	statusX := http.StatusBadRequest
	if len(status) > 0 {
		statusX = status[0]
	}

	// httprequest, _ := httputil.DumpRequest(ctx.Request, false)
	// goErr := errors.Wrap(err, 3)
	// reset := string([]byte{27, 91, 48, 109})
	// ctx.Logger.Errorf("[Nice ctx.Fail] error:\n\n%s%s\n\n%s%s", httprequest, goErr.Error(), goErr.Stack(), reset)

	ctx.Logger.Infof("[ctx.Fail] error: %s", err)

	if ok := ctx.Debug().IsDebugMode(); ok {
		fmt.Println("[ctx.Fail] error stack: \n", string(rd.Stack())+"\n")
	}

	ctx.JSON(statusX, map[string]any{
		"code":    code,
		"message": message,
	})
}

// FailWithError writes the given error with code-message-result specification to the response.
func (ctx *Context) FailWithError(err HTTPError) {
	ctx.Fail(err.Raw(), err.Code(), err.Message(), err.Status())
}

// Redirect redirects the request to the given URL.
func (ctx *Context) Redirect(url string, status ...int) {
	ctx.SetLocation(url)

	code := http.StatusFound
	if len(status) == 1 && status[0] != 0 {
		code = status[0]
	}

	ctx.Status(code)
}

// RedirectTemporary redirects the request temporarily to the given URL.
func (ctx *Context) RedirectTemporary(url string) {
	ctx.Redirect(url, http.StatusFound)
}

// RedirectPermanent redirects the request permanently to the given URL.
func (ctx *Context) RedirectPermanent(url string) {
	ctx.Redirect(url, http.StatusMovedPermanently)
}

// RedirectSeeOther redirects the request to the given URL.
func (ctx *Context) RedirectSeeOther(url string) {
	ctx.Redirect(url, http.StatusSeeOther)
}

// RedirectTemporaryWithOriginMethodAndBody redirects the request temporarily to the given URL with the origin method.
func (ctx *Context) RedirectTemporaryWithOriginMethodAndBody(url string) {
	ctx.Redirect(url, http.StatusTemporaryRedirect)
}

// RedirectPermanentWithOriginMethodAndBody redirects the request permanently to the given URL with the origin method.
func (ctx *Context) RedirectPermanentWithOriginMethodAndBody(url string) {
	ctx.Redirect(url, http.StatusPermanentRedirect)
}

// Protocol returns the protocol, usally http or https
func (ctx *Context) Protocol() string {
	return ctx.Request.URL.Scheme
}

// Host gets the host from HTTP Header.
// format: `host:port`
func (ctx *Context) Host() string {
	return ctx.Request.Host
}

// Hostname gets the hostname from HTTP Header.
// format: `hostname`
func (ctx *Context) Hostname() string {
	hostname := ctx.Request.URL.Hostname()
	if hostname != "" {
		return hostname
	}

	hostname, _ = utils.SplitHostPort(ctx.Request.Host)
	return hostname
}

// URL is http.Request.RequestURI.
func (ctx *Context) URL() string {
	return ctx.Request.RequestURI
}

// IP gets the ip from X-Forwarded-For or X-Real-IP or RemoteIP.
// RemoteIP parses the IP from Request.RemoteAddr, normializes and returns the IP (without the port).
func (ctx *Context) IP() string {
	if xForwardedFor := ctx.Header().Get(headers.XForwardedFor); xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	if xRealIP := ctx.Header().Get(headers.XRealIP); xRealIP != "" {
		return xRealIP
	}

	ip, _, err := net.SplitHostPort(strings.TrimSpace(ctx.Request.RemoteAddr))
	if err != nil {
		return ""
	}

	return ip
}

// IPs gets the ip from X-Forwarded-For or X-Real-IP or RemoteIP.
// RemoteIP parses the IP from Request.RemoteAddr, normializes and returns the IP (without the port).
func (ctx *Context) IPs() []string {
	if xForwardedFor := ctx.Header().Get(headers.XForwardedFor); xForwardedFor != "" {
		return strings.Split(xForwardedFor, ",")
	}

	return []string{ctx.IP()}
}

// ClientIP is the client ip.
func (ctx *Context) ClientIP() string {
	return ctx.IP()
}

// Headers gets all headers.
func (ctx *Context) Headers() *safe.Map[string, any] {
	headers := safe.NewMap[string, any]()

	for key, values := range ctx.Request.Header {
		headers.Set(key, values[0])
	}

	return headers
}

// Queries gets all queries.
func (ctx *Context) Queries() *safe.Map[string, any] {
	queries := safe.NewMap[string, any]()

	for key, values := range ctx.Request.URL.Query() {
		queries.Set(key, values[0])
	}

	return queries
}

// Forms gets all forms.
func (ctx *Context) Forms() (*safe.Map[string, any], error) {
	forms := safe.NewMap[string, any]()

	if err := ctx.Request.ParseForm(); err != nil {
		// http: request body too large
		if err.Error() == "http: request body too large" {
			return nil, errors.New("request body too large")
		}

		return nil, err
	}

	for key, values := range ctx.Request.Form {
		forms.Set(key, values[0])
	}

	return forms, nil
}

// Params gets all params.
func (ctx *Context) Params() *safe.Map[string, any] {
	m := safe.NewMap[string, any]()
	if ctx.param == nil {
		return m
	}

	for k, v := range ctx.param.Iterator() {
		m.Set(k, v)
	}

	return m
}

// Bodies gets all bodies.
func (ctx *Context) Bodies() map[string]any {
	var bodies map[string]any

	if bytes, err := ctx.BodyBytes(); err == nil {
		if err := json.Unmarshal(bytes, &bodies); err == nil {
			return bodies
		}
	}

	return nil
}

// Cookies gets all cookies.
func (ctx *Context) Cookies() map[string]string {
	cookies := map[string]string{}

	for _, cookie := range ctx.Request.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	return cookies
}

// Files gets all files.
func (ctx *Context) Files() map[string]*multipart.FileHeader {
	if err := ctx.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil
	}

	if ctx.Request.MultipartForm == nil || ctx.Request.MultipartForm.File == nil {
		return nil
	}

	files := map[string]*multipart.FileHeader{}
	for key, file := range ctx.Request.MultipartForm.File {
		if len(file) > 0 {
			files[key] = file[0]
		}
	}

	return files
}

// File gets the file by key.
func (ctx *Context) File(key string) (multipart.File, *multipart.FileHeader, error) {
	return ctx.Request.FormFile(key)
}

// Stream get the body stream.
func (ctx *Context) Stream() io.ReadCloser {
	return ctx.Request.Body
}

// GetRawData returns stream data.
// Align to gin framework.
func (ctx *Context) GetRawData() ([]byte, error) {
	return ioutil.ReadAll(ctx.Request.Body)
}

// BindJSON binds the request body into the given struct.
func (ctx *Context) BindJSON(obj interface{}) (err error) {
	if !strings.Contains(ctx.Header().Get("Content-Type"), "application/json") {
		return errors.New("[BindJSON] content-type is not json")
	}

	if ctx.Request.Body == nil {
		return errors.New("invalid request")
	}

	if ctx.Env().Get("DEBUG_ZOOX_REQUEST_BODY") != "" {
		// refernece: golang复用http.request.body - https://zhuanlan.zhihu.com/p/47313038
		_, err = ctx.CloneBody()
		if err != nil {
			return fmt.Errorf("failed to read request body: %v", err)
		}

		ctx.Logger.Infof("[debug][ctx.BindJSON] body: %s", ctx.bodyBytes)
	}

	if err := json.NewDecoder(ctx.Request.Body).Decode(obj); err != nil {
		// @TODO allow empty body
		if err == io.EOF {
			return nil
		}

		// request body too large
		if err.Error() == "http: request body too large" {
			return errors.New("request body too large")
		}

		return err
	}

	return nil
}

// BindYAML binds the request body into the given struct.
func (ctx *Context) BindYAML(obj interface{}) (err error) {
	if ctx.Request.Body == nil {
		return errors.New("invalid request")
	}

	if ctx.Debug().IsDebugMode() {
		// refernece: golang复用http.request.body - https://zhuanlan.zhihu.com/p/47313038
		_, err = ctx.CloneBody()
		if err != nil {
			return fmt.Errorf("failed to read request body: %v", err)
		}

		ctx.Logger.Infof("[debug][ctx.BindYAML] body: %v", ctx.bodyBytes)
	}

	return yaml.NewDecoder(ctx.Request.Body).Decode(obj)
}

// BindForm binds the query into the given struct.
func (ctx *Context) BindForm(obj interface{}) error {
	forms, err := ctx.Forms()
	if err != nil {
		return err
	}

	if ctx.Debug().IsDebugMode() {
		ctx.Logger.Infof("[debug][ctx.BindForm]")
		for k, v := range forms.ToMap() {
			ctx.Logger.Infof("[debug][ctx.BindForm][detail] %s = %s", k, v)
		}
	}

	return tag.New("form", datasource.GetterToDataSource(forms)).Decode(obj)
}

// BindParams binds the params into the given struct.
func (ctx *Context) BindParams(obj interface{}) error {
	params := ctx.Params()
	if ctx.Debug().IsDebugMode() {
		ctx.Logger.Infof("[debug][ctx.BindParams]")
		for k, v := range params.ToMap() {
			ctx.Logger.Infof("[debug][ctx.BindParams][detail] %s = %s", k, v)
		}
	}

	return tag.New("param", datasource.GetterToDataSource(params)).Decode(obj)
}

// BindHeader binds the header into the given struct.
func (ctx *Context) BindHeader(obj interface{}) error {
	headers := ctx.Headers()
	if ctx.Debug().IsDebugMode() {
		ctx.Logger.Infof("[debug][ctx.BindHeader]")
		for k, v := range headers.ToMap() {
			ctx.Logger.Infof("[debug][ctx.BindHeader][detail] %s = %s", k, v)
		}
	}

	return tag.New("header", datasource.GetterToDataSource(headers)).Decode(obj)
}

// BindQuery binds the query into the given struct.
func (ctx *Context) BindQuery(obj interface{}) error {
	queries := ctx.Queries()
	if ctx.Debug().IsDebugMode() {
		ctx.Logger.Infof("[debug][ctx.BindQuery]")
		for k, v := range queries.ToMap() {
			ctx.Logger.Infof("[debug][ctx.BindQuery][detail] %s = %s", k, v)
		}
	}

	return tag.New("query", datasource.GetterToDataSource(queries)).Decode(obj)
}

// BindBody binds the body into the given struct.
func (ctx *Context) BindBody(obj interface{}) error {
	data := ctx.Bodies()
	if ctx.Debug().IsDebugMode() {
		ctx.Logger.Infof("[debug][ctx.BindBody]")
		for k, v := range data {
			ctx.Logger.Infof("[debug][ctx.BindBody][detail] %s = %v", k, v)
		}
	}

	return tag.New("body", datasource.NewMapDataSource(data)).Decode(obj)
}

// SaveFile saves the file to the given path.
func (ctx *Context) SaveFile(key, path string) error {
	src, _, err := ctx.Request.FormFile(key)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

// AcceptJSON returns true if the request accepts json.
func (ctx *Context) AcceptJSON() bool {
	accept := ctx.Header().Get(headers.Accept)
	// for curl
	if accept == "*/*" {
		return true
	}

	return strings.Contains(accept, "application/json")
}

// AcceptHTML returns true if the request accepts html.
func (ctx *Context) AcceptHTML() bool {
	return strings.Contains(ctx.Header().Get(headers.Accept), "text/html")
}

// Cache returns the cache of the application.
func (ctx *Context) Cache() cache.Cache {
	ctx.once.cache.Do(func() {
		ctx.cache = ctx.App.Cache()
	})

	return ctx.cache
}

// Cron returns the cache of the application.
func (ctx *Context) Cron() cron.Cron {
	ctx.once.cron.Do(func() {
		ctx.cron = ctx.App.Cron()
	})

	return ctx.cron
}

// JobQueue returns the queue of the application.
func (ctx *Context) JobQueue() jobqueue.JobQueue {
	ctx.once.queue.Do(func() {
		ctx.queue = ctx.App.JobQueue()
	})

	return ctx.queue
}

// I18n returns the i18n of the application.
func (ctx *Context) I18n() i18n.I18n {
	ctx.once.i18n.Do(func() {
		ctx.i18n = ctx.App.I18n()
	})

	return ctx.i18n
}

// Debug returns the debug of the app.
func (ctx *Context) Debug() debug.Debug {
	ctx.once.debug.Do(func() {
		ctx.debug = ctx.App.Debug()
	})

	return ctx.debug
}

// Env returns the env of the
func (ctx *Context) Env() env.Env {
	ctx.once.env.Do(func() {
		ctx.env = ctx.App.Env()
	})

	return ctx.env
}

// State returns the state of the
func (ctx *Context) State() state.State {
	ctx.once.state.Do(func() {
		ctx.state = state.New()
	})

	return ctx.state
}

// User returns the user of the
func (ctx *Context) User() user.User {
	ctx.once.user.Do(func() {
		ctx.user = user.New()
	})

	return ctx.user
}

// Cmd returns the cmd of the request.
func (ctx *Context) Cmd() cmd.Cmd {
	ctx.once.cmd.Do(func() {
		ctx.cmd = cmd.New(ctx.Context())
	})

	return ctx.cmd
}

// Cookie returns the cookie of the request.
func (ctx *Context) Cookie() cookie.Cookie {
	ctx.once.cookie.Do(func() {
		ctx.cookie = cookie.New(
			ctx.Writer,
			ctx.Request,
		)
	})

	return ctx.cookie
}

// Session returns the session of the request.
func (ctx *Context) Session() session.Session {
	ctx.once.session.Do(func() {
		secretKey := ctx.App.Config.SecretKey
		if secretKey == "" {
			secretKey = "go-zoox_" + random.String(24)
		}

		ctx.session = session.New(ctx.Cookie(), secretKey, &ctx.App.Config.Session)
	})

	return ctx.session
}

// Jwt returns the jwt of the request.
func (ctx *Context) Jwt() jwt.Jwt {
	ctx.once.jwt.Do(func() {
		secretKey := ctx.App.Config.SecretKey
		if secretKey == "" {
			secretKey = "go-zoox_" + random.String(24)
		}

		ctx.jwt = jwt.New(secretKey)
	})

	return ctx.jwt
}

// RequestID returns the request id of the request.
func (ctx *Context) RequestID() string {
	return ctx.requestID
}

// Fetch is the context request utils, based on go-zoox/fetch.
func (ctx *Context) Fetch() *fetch.Fetch {
	return fetch.New()
}

// Proxy customize the request to proxy the backend services.
func (ctx *Context) Proxy(target string, cfg ...*proxy.SingleHostConfig) {
	WrapH(proxy.NewSingleHost(target, cfg...))(ctx)
}

// CloneBody clones the body of the request, should be used carefully.
func (ctx *Context) CloneBody() (body io.ReadCloser, err error) {
	if ctx.bodyBytes == nil {
		// refernece: golang复用http.request.body - https://zhuanlan.zhihu.com/p/47313038
		ctx.bodyBytes, err = ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}

		// recovery to request body
		ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(ctx.bodyBytes))
	}

	return ioutil.NopCloser(bytes.NewBuffer(ctx.bodyBytes)), nil
}

// BodyBytes reads all bodies as string.
func (ctx *Context) BodyBytes() ([]byte, error) {
	bytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// PubSub is the pubsub.
func (ctx *Context) PubSub() pubsub.PubSub {
	ctx.once.pubsub.Do(func() {
		ctx.pubsub = pubsub.New(ctx.Context(), ctx.App.PubSub())
	})

	return ctx.pubsub
}

// MQ is the mq.
func (ctx *Context) MQ() mq.MQ {
	ctx.once.mq.Do(func() {
		ctx.mq = mq.New(ctx.Context(), ctx.App.MQ())
	})

	return ctx.mq
}

// Concurrency creates a concurrency.
func (ctx *Context) Concurrency(limit int) *concurrency.Concurrency {
	return concurrency.New(limit)
}
