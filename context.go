package zoox

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/go-zoox/core-utils/safe"
	"github.com/go-zoox/fetch"
	"github.com/go-zoox/logger"
	"github.com/go-zoox/tag"
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
	param *Param
	query *Query
	form  *Form

	// response
	StatusCode int
	//
	cookie  *Cookie
	session *Session
	//
	cache *Cache
	cron  *Cron
	queue *Queue
	//
	env   *Env
	debug *Debug
	// middleware
	handlers []HandlerFunc
	index    int
	//
	App *Application
	//
	Logger *logger.Logger
	//
	//
	state *State
	user  *User
	// request id
	requestID string
}

func newContext(app *Application, w http.ResponseWriter, req *http.Request) *Context {
	// path := strings.TrimSuffix(req.URL.Path, "/")
	path := req.URL.Path
	// path := req.URL.Path
	// if !strings.HasSuffix(path, "/") {
	// 	path += "/"
	// }

	ctx := &Context{
		App:        app,
		Writer:     newResponseWriter(w),
		Request:    req,
		Method:     req.Method,
		Path:       path,
		StatusCode: 404,
		index:      -1,
	}

	ctx.requestID = ctx.Get(RequestIDHeader)
	if ctx.requestID == "" {
		ctx.requestID = GenerateRequestID()
	}

	ctx.Logger = logger.New(&logger.Options{
		Level: app.LogLevel,
	})

	ctx.Writer.setContext(ctx)

	return ctx
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
func (ctx *Context) Query() *Query {
	if ctx.query == nil {
		ctx.query = newQuery(ctx)
	}

	return ctx.query
}

// Param returns the named URL parameter value if it exists.
func (ctx *Context) Param() *Param {
	return ctx.param
}

// Form returns the form data from POST or PUT request body.
func (ctx *Context) Form() *Form {
	if ctx.form == nil {
		ctx.form = newForm(ctx)
	}

	return ctx.form
}

// Status sets the HTTP response status code.
func (ctx *Context) Status(status int) {
	ctx.StatusCode = status
	ctx.Writer.WriteHeader(status)
}

// Get alias for ctx.Header.
func (ctx *Context) Get(key string) string {
	return ctx.Header(key)
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

// BasicAuth returns the user/password pair for Basic Authentication.
func (ctx *Context) BasicAuth() (string, string, bool) {
	return ctx.Request.BasicAuth()
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
	ctx.SetHeader("content-type", "application/json")
	encoder := json.NewEncoder(ctx.Writer)
	if err := encoder.Encode(obj); err != nil {
		ctx.Error(http.StatusInternalServerError, err.Error())
	}
}

// Data writes some data into the body stream and updates the HTTP code.
// Align to gin framework.
func (ctx *Context) Data(status int, contentType string, data []byte) {
	ctx.Status(status)
	ctx.SetHeader("content-type", contentType)
	ctx.Write(data)
}

// HTML renders the given template with the given data and writes the result
func (ctx *Context) HTML(code int, name string, data interface{}) {
	ctx.Status(code)
	ctx.SetHeader("content-type", "text/html")
	if err := ctx.App.templates.ExecuteTemplate(ctx.Writer, name, data); err != nil {
		ctx.Fail(err, http.StatusInternalServerError, err.Error())
	}
}

// Render renders a template with data and writes the result to the response.
func (ctx *Context) Render(code int, name string, data interface{}) {
	ctx.HTML(code, name, data)
}

// Error writes the given error to the response.
// Use for system errors
//	1. Internal server error
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

	// ctx.Logger.Error("[Fail] %s", err)
	fmt.Println("[Fail]", err)

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

// Host gets the host from HTTP Header.
// format: `host:port`
func (ctx *Context) Host() string {
	return ctx.Request.Host
}

// URL is http.Request.RequestURI.
func (ctx *Context) URL() string {
	return ctx.Request.RequestURI
}

// IP gets the ip from X-Forwarded-For or X-Real-IP or RemoteAddr.
func (ctx *Context) IP() string {
	if xForwardedFor := ctx.Header("X-Forwarded-For"); xForwardedFor != "" {
		return strings.Split(xForwardedFor, ",")[0]
	}

	if xRealIP := ctx.Header("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	return ctx.Request.RemoteAddr
}

// Header gets the header value by key.
func (ctx *Context) Header(key string) string {
	return ctx.Request.Header.Get(key)
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

	if bytes, err := io.ReadAll(ctx.Request.Body); err == nil {
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
func (ctx *Context) BindJSON(obj interface{}) error {
	if !strings.Contains(ctx.Get("Content-Type"), "application/json") {
		return errors.New("[BindJSON] content-type is not json")
	}

	if ctx.Request.Body == nil {
		return errors.New("invalid request")
	}

	return json.NewDecoder(ctx.Request.Body).Decode(obj)
}

// BindYAML binds the request body into the given struct.
func (ctx *Context) BindYAML(obj interface{}) error {
	if ctx.Request.Body == nil {
		return errors.New("invalid request")
	}

	return yaml.NewDecoder(ctx.Request.Body).Decode(obj)
}

// BindForm binds the query into the given struct.
func (ctx *Context) BindForm(obj interface{}) error {
	return tag.New("form", ctx.Forms()).Decode(obj)
}

// BindParams binds the params into the given struct.
func (ctx *Context) BindParams(obj interface{}) error {
	return tag.New("param", ctx.Params()).Decode(obj)
}

// BindHeader binds the header into the given struct.
func (ctx *Context) BindHeader(obj interface{}) error {
	return tag.New("header", ctx.Headers()).Decode(obj)
}

// BindQuery binds the query into the given struct.
func (ctx *Context) BindQuery(obj interface{}) error {
	return tag.New("query", ctx.Queries()).Decode(obj)
}

// // BindBody binds the body into the given struct.
// func (ctx *Context) BindBody(obj interface{}) error {
// 	return tag.New("body", ctx.Bodies()).Decode(obj)
// }

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
	accept := ctx.Header("Accept")
	// for curl
	if accept == "*/*" {
		return true
	}

	return strings.Contains(accept, "application/json")
}

// Origin returns the origin of the request.
func (ctx *Context) Origin() string {
	return ctx.Get("Origin")
}

// Cache returns the cache of the application.
func (ctx *Context) Cache() *Cache {
	if ctx.cache == nil {
		ctx.cache = ctx.App.Cache()
	}

	return ctx.cache
}

// Cron returns the cache of the application.
func (ctx *Context) Cron() *Cron {
	if ctx.cron == nil {
		ctx.cron = ctx.App.Cron()
	}

	return ctx.cron
}

// Queue returns the queue of the application.
func (ctx *Context) Queue() *Queue {
	if ctx.queue == nil {
		ctx.queue = ctx.App.Queue()
	}

	return ctx.queue
}

// Debug returns the debug of the app.
func (ctx *Context) Debug() *Debug {
	if ctx.debug == nil {
		ctx.debug = ctx.App.Debug()
	}

	return ctx.debug
}

// Env returns the env of the
func (ctx *Context) Env() *Env {
	if ctx.env == nil {
		ctx.env = ctx.App.Env
	}

	return ctx.env
}

// State returns the state of the
func (ctx *Context) State() *State {
	if ctx.state == nil {
		ctx.state = newState()
	}

	return ctx.state
}

// User returns the user of the
func (ctx *Context) User() *User {
	if ctx.user == nil {
		ctx.user = newUser()
	}

	return ctx.user
}

// Cookie returns the cookie of the request.
func (ctx *Context) Cookie() *Cookie {
	if ctx.cookie == nil {
		ctx.cookie = newCookie(ctx)
	}

	return ctx.cookie
}

// Session returns the session of the request.
func (ctx *Context) Session() *Session {
	if ctx.session == nil {
		ctx.session = newSession(ctx)
	}

	return ctx.session
}

// RequestID returns the request id of the request.
func (ctx *Context) RequestID() string {
	return ctx.requestID
}

// Fetch is the context request utils, based on go-zoox/fetch.
func (ctx *Context) Fetch() *fetch.Fetch {
	return fetch.New()
}
