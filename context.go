package zoox

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	Params map[string]string
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
	if value, ok := ctx.Params[key]; ok {
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
