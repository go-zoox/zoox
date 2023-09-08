package zoox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"

	"time"

	"github.com/go-zoox/cache"
	"github.com/go-zoox/fs"
	"github.com/go-zoox/i18n"
	"github.com/go-zoox/proxy"
	"github.com/go-zoox/pubsub"
	"github.com/go-zoox/tag/datasource"
	"github.com/go-zoox/zoox/components/application/cron"
	"github.com/go-zoox/zoox/components/application/debug"
	"github.com/go-zoox/zoox/components/application/env"
	"github.com/go-zoox/zoox/components/application/jobqueue"
	"github.com/go-zoox/zoox/components/context/body"
	"github.com/go-zoox/zoox/components/context/form"
	"github.com/go-zoox/zoox/components/context/param"
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
	"gopkg.in/yaml.v3"
)

// Context is the request context
type Context struct {
	// origin objects
	Writer  ResponseWriter
	Request *http.Request
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
	}
}

func newContext(app *Application, w http.ResponseWriter, req *http.Request) *Context {
	// path := strings.TrimSuffix(req.URL.Path, "/")
	path := req.URL.Path
	// path := req.URL.Path
	// if !strings.HasSuffix(path, "/") {
	// 	path += "/"
	// }

	ctx := &Context{
		App:     app,
		Writer:  newResponseWriter(w),
		Request: req,
		Method:  req.Method,
		Path:    path,
		//
		index: -1,
	}

	ctx.requestID = ctx.Get(utils.RequestIDHeader)
	if ctx.requestID == "" {
		ctx.requestID = utils.GenerateRequestID()
	}

	ctx.Logger = logger.New(&logger.Options{
		Level: app.Config.LogLevel,
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
	return ctx.Get(headers.Authorization)
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
	return ctx.Get(headers.Accept)
}

// AcceptLanguage returns the request accept header.
func (ctx *Context) AcceptLanguage() string {
	return ctx.Get(headers.AcceptLanguage)
}

// AcceptEncoding returns the request accept header.
func (ctx *Context) AcceptEncoding() string {
	return ctx.Get(headers.AcceptEncoding)
}

// Connection return the request connection header.
func (ctx *Context) Connection() string {
	return ctx.Get(headers.Connection)
}

// UserAgent return the request user-agent header.
func (ctx *Context) UserAgent() string {
	return ctx.Get(headers.UserAgent)
}

// XForwardedFor return the request x-forwarded-for header.
func (ctx *Context) XForwardedFor() string {
	return ctx.Get(headers.XForwardedFor)
}

// XForwardedProto return the request x-forwarded-proto header.
func (ctx *Context) XForwardedProto() string {
	return ctx.Get(headers.XForwardedProto)
}

// XForwardedHost return the request x-forwarded-host header.
func (ctx *Context) XForwardedHost() string {
	return ctx.Get(headers.XForwardedHost)
}

// XForwardedPort return the request x-forwarded-port header.
func (ctx *Context) XForwardedPort() string {
	return ctx.Get(headers.XForwardedPort)
}

// XRealIP return the request x-real-ip header.
func (ctx *Context) XRealIP() string {
	return ctx.Get(headers.XRealIP)
}

// Upgrade return the request upgrade header.
func (ctx *Context) Upgrade() string {
	return ctx.Get(headers.Upgrade)
}

// Origin returns the origin of the request.
func (ctx *Context) Origin() string {
	return ctx.Get(headers.Origin)
}

// Referrer returns the referrer of the request.
func (ctx *Context) Referrer() string {
	return ctx.Get(headers.Referrer)
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
		ctx.Error(http.StatusInternalServerError, err.Error())
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
func (ctx *Context) HTML(status int, html string) {
	ctx.SetHeader(headers.ContentType, "text/html")
	ctx.String(status, html)
}

// Template renders the given template with the given data and writes the result
func (ctx *Context) Template(status int, name string, data interface{}) {
	if ctx.App.templates == nil {
		ctx.Error(http.StatusInternalServerError, "templates is not initialized, please use app.SetTemplates() to initialize")
		return
	}

	ctx.Status(status)
	ctx.SetHeader(headers.ContentType, "text/html")
	if err := ctx.App.templates.ExecuteTemplate(ctx.Writer, name, data); err != nil {
		ctx.Fail(err, http.StatusInternalServerError, err.Error())
	}
}

// Render renders a template with data and writes the result to the response.
func (ctx *Context) Render(status int, name string, data interface{}) {
	ctx.Template(status, name, data)
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
			"code":      400,
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

	funcName := "unknown"
	// get panic error occurred file and line
	pc, filepath, line, ok := runtime.Caller(2)
	if ok {
		filepath = filepath[len(fs.CurrentDir())+1:]
		funcName = runtime.FuncForPC(pc).Name()
		funcNameParts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		if len(funcNameParts) > 0 {
			funcName = funcNameParts[len(funcNameParts)-1]
		}
	}

	ctx.Logger.Errorf("[fail][%s:%d,%s][%s %s] %s", filepath, line, funcName, ctx.Method, ctx.Path, err)

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
	ctx.SetHeader("location", url)

	code := http.StatusFound
	if len(status) == 1 && status[0] != 0 {
		code = status[0]
	}

	ctx.Status(code)
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
	if xForwardedFor := ctx.Get(headers.XForwardedFor); xForwardedFor != "" {
		parts := strings.Split(xForwardedFor, ",")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	if xRealIP := ctx.Get(headers.XRealIP); xRealIP != "" {
		return xRealIP
	}

	ip, _, err := net.SplitHostPort(strings.TrimSpace(ctx.Request.RemoteAddr))
	if err != nil {
		return ""
	}

	return ip
}

// ClientIP is the client ip.
func (ctx *Context) ClientIP() string {
	return ctx.IP()
}

// Headers gets all headers.
func (ctx *Context) Headers() *safe.Map {
	headers := safe.NewMap()

	for key, values := range ctx.Request.Header {
		headers.Set(key, values[0])
	}

	return headers
}

// Queries gets all queries.
func (ctx *Context) Queries() *safe.Map {
	queries := safe.NewMap()

	for key, values := range ctx.Request.URL.Query() {
		queries.Set(key, values[0])
	}

	return queries
}

// Forms gets all forms.
func (ctx *Context) Forms() *safe.Map {
	forms := safe.NewMap()

	if err := ctx.Request.ParseForm(); err != nil {
		return forms
	}

	for key, values := range ctx.Request.Form {
		forms.Set(key, values[0])
	}

	return forms
}

// Params gets all params.
func (ctx *Context) Params() *safe.Map {
	m := safe.NewMap()
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
func (ctx *Context) File(key string) (multipart.File, *multipart.FileHeader) {
	if file, header, err := ctx.Request.FormFile(key); err == nil {
		return file, header
	}

	return nil, nil
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
	if !strings.Contains(ctx.Get("Content-Type"), "application/json") {
		return errors.New("[BindJSON] content-type is not json")
	}

	if ctx.Request.Body == nil {
		return errors.New("invalid request")
	}

	if ctx.Debug().IsDebugMode() {
		// refernece: golang复用http.request.body - https://zhuanlan.zhihu.com/p/47313038
		_, err = ctx.CloneBody()
		if err != nil {
			return fmt.Errorf("failed to read request body: %v", err)
		}

		ctx.Logger.Infof("[debug][ctx.BindJSON] body: %v", ctx.bodyBytes)
	}

	return json.NewDecoder(ctx.Request.Body).Decode(obj)
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
	forms := ctx.Forms()
	if ctx.Debug().IsDebugMode() {
		ctx.Logger.Infof("[debug][ctx.BindForm]")
		for k, v := range forms.ToMap() {
			ctx.Logger.Infof("[debug][ctx.BindForm][detail] %s = %s", k, v)
		}
	}

	return tag.New("form", forms).Decode(obj)
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

	return tag.New("param", params).Decode(obj)
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

	return tag.New("header", headers).Decode(obj)
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

	return tag.New("query", queries).Decode(obj)
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
	accept := ctx.Get(headers.Accept)
	// for curl
	if accept == "*/*" {
		return true
	}

	return strings.Contains(accept, "application/json")
}

// AcceptHTML returns true if the request accepts html.
func (ctx *Context) AcceptHTML() bool {
	return strings.Contains(ctx.Get(headers.Accept), "text/html")
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
		ctx.env = ctx.App.Env
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

// Publish publishes the message to the pubsub.
func (ctx *Context) Publish(msg *pubsub.Message) error {
	return ctx.App.PubSub().Publish(ctx.Context(), msg)
}

// Subscribe subscribes the topic with the handler.
func (ctx *Context) Subscribe(topic string, handler pubsub.Handler) error {
	return ctx.App.PubSub().Subscribe(ctx.Context(), topic, handler)
}

// Concurrency creates a concurrency.
func (ctx *Context) Concurrency(limit int) *concurrency.Concurrency {
	return concurrency.New(limit)
}
