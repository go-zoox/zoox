package zoox

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/go-yaml/yaml"
)

// Context is the request context
type Context struct {
	// origin objects
	Writer  http.ResponseWriter
	Request *http.Request
	// request
	Method string
	Path   string
	//
	params map[string]string
	// response
	StatusCode int
	// middleware
	handlers []HandlerFunc
	index    int
	//
	app *Application
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	// path := strings.TrimSuffix(req.URL.Path, "/")
	path := req.URL.Path
	// path := req.URL.Path
	// if !strings.HasSuffix(path, "/") {
	// 	path += "/"
	// }

	return &Context{
		Writer:     w,
		Request:    req,
		Method:     req.Method,
		Path:       path,
		StatusCode: 200,
		index:      -1,
	}
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
func (ctx *Context) Query(key string) string {
	return ctx.Request.URL.Query().Get(key)
}

// Param returns the named URL parameter value if it exists.
func (ctx *Context) Param(key string) string {
	if value, ok := ctx.params[key]; ok {
		return value
	}

	return ""
}

// PostForm returns the form data from POST or PUT request body.
func (ctx *Context) PostForm(key string) string {
	return ctx.Request.FormValue(key)
}

// Status sets the HTTP response status code.
func (ctx *Context) Status(status int) {
	ctx.StatusCode = status
	ctx.Writer.WriteHeader(status)
}

// SetHeader sets a header in the response.
func (ctx *Context) SetHeader(key string, value string) {
	ctx.Writer.Header().Set(key, value)
}

// SetCookie sets a cookie with the given name and value.
func (ctx *Context) SetCookie(name string, value string, maxAge time.Duration) {
	expires := time.Now().Add(maxAge)

	ctx.SetHeader(
		"Set-Cookie",
		fmt.Sprintf("%s=%s; path=/; expires=%s; httponly", name, value, expires),
	)
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

// HTML renders the given template with the given data and writes the result
func (ctx *Context) HTML(code int, name string, data interface{}) {
	ctx.Status(code)
	ctx.SetHeader("content-type", "text/html")
	if err := ctx.app.templates.ExecuteTemplate(ctx.Writer, name, data); err != nil {
		ctx.Fail(http.StatusInternalServerError, err.Error())
	}
}

// Render renders a template with data and writes the result to the response.
func (ctx *Context) Render(code int, name string, data interface{}) {
	ctx.HTML(code, name, data)
}

// Error writes the given error to the response.
func (ctx *Context) Error(status int, message string) {
	ctx.Status(status)
	ctx.Write([]byte(message))
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
func (ctx *Context) Fail(code int, message string) {
	ctx.JSON(http.StatusBadRequest, map[string]any{
		"code":    code,
		"message": message,
	})
}

// Redirect redirects the request to the given URL.
func (ctx *Context) Redirect(url string, status ...int) {
	code := http.StatusFound
	if len(status) == 1 && status[0] != 0 {
		code = status[0]
	}

	ctx.Status(code)
	ctx.SetHeader("location", url)
}

// Header gets the header value by key.
func (ctx *Context) Header(key string) string {
	return ctx.Request.Header.Get(key)
}

// Headers gets all headers.
func (ctx *Context) Headers() map[string]string {
	headers := map[string]string{}

	for key, values := range ctx.Request.Header {
		headers[key] = values[0]
	}

	return headers
}

// Queries gets all queries.
func (ctx *Context) Queries() map[string]string {
	queries := map[string]string{}

	for key, values := range ctx.Request.URL.Query() {
		queries[key] = values[0]
	}

	return queries
}

// Forms gets all forms.
func (ctx *Context) Forms() map[string]string {
	forms := map[string]string{}

	if err := ctx.Request.ParseForm(); err != nil {
		return forms
	}

	for key, values := range ctx.Request.Form {
		forms[key] = values[0]
	}

	return forms
}

// Params gets all params.
func (ctx *Context) Params() map[string]string {
	return ctx.params
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

// Cookie gets the cookie value by key.
func (ctx *Context) Cookie(key string) string {
	cookie, err := ctx.Request.Cookie(key)
	if err != nil {
		return ""
	}

	return cookie.Value
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
func (ctx *Context) File(key string) multipart.File {
	if file, _, err := ctx.Request.FormFile(key); err == nil {
		return file
	} else {
		return nil
	}
}

// Stream get the body stream.
func (ctx *Context) Stream() io.ReadCloser {
	return ctx.Request.Body
}

// BindJSON binds the request body into the given struct.
func (ctx *Context) BindJSON(obj interface{}) error {
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
	forms := ctx.Forms()
	jf, _ := json.Marshal(forms)
	return json.Unmarshal(jf, obj)
}

// BindParams binds the params into the given struct.
func (ctx *Context) BindParams(obj interface{}) error {
	params := ctx.Params()
	jf, _ := json.Marshal(params)
	return json.Unmarshal(jf, obj)
}

// BindHeader binds the header into the given struct.
func (ctx *Context) BindHeader(obj interface{}) error {
	headers := ctx.Headers()
	jf, _ := json.Marshal(headers)
	return json.Unmarshal(jf, obj)
}

// BindQuery binds the query into the given struct.
func (ctx *Context) BindQuery(obj interface{}) error {
	queries := ctx.Queries()
	jf, _ := json.Marshal(queries)
	return json.Unmarshal(jf, obj)
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
