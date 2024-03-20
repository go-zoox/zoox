package zoox

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/go-zoox/headers"
)

// TemplateConfig is the template config.
type TemplateConfig struct {
	// ContentType is the template content type, default is "text/plain".
	ContentType string `json:"content_type"`
	// Name is the template name.
	Name string `json:"name"`
	// Content is the template content.
	Content string `json:"content"`
	// Data is the template data.
	Data any `json:"data"`
}

// TemplateOption is the template option.
type TemplateOption func(*TemplateConfig)

// Template renders the given template with the given data and writes the result
func (ctx *Context) Template(status int, opts ...TemplateOption) {
	cfg := &TemplateConfig{
		ContentType: "text/plain",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// if content is not empty, use content as template
	if cfg.Content != "" {
		tmpl, err := template.New("example").Parse(cfg.Content)
		if err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}

		var output string
		buf := &bytes.Buffer{}
		if err = tmpl.Execute(buf, cfg.Data); err != nil {
			ctx.Error(http.StatusInternalServerError, err.Error())
			return
		}
		output = buf.String()

		ctx.SetHeader(headers.ContentType, cfg.ContentType)
		ctx.String(status, output)
		return
	}

	// if name is not empty, use template file by name
	if ctx.App.templates == nil {
		ctx.Error(http.StatusInternalServerError, "templates is not initialized, please use app.SetTemplates() to initialize")
		return
	}

	ctx.Status(status)
	ctx.SetHeader(headers.ContentType, cfg.ContentType)
	if err := ctx.App.templates.ExecuteTemplate(ctx.Writer, cfg.Name, cfg.Data); err != nil {
		ctx.Fail(err, http.StatusInternalServerError, err.Error())
	}
}
