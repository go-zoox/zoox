# Tutorial 06: Template Engine

## 📖 Overview

Learn how to work with templates in Zoox for building dynamic web applications. This tutorial covers template setup, inheritance, data binding, and custom helpers.

## 🎯 Learning Objectives

- Set up and configure template engines
- Create template layouts and inheritance
- Bind dynamic data to templates
- Build custom template helpers
- Implement template caching and optimization

## 📋 Prerequisites

- Completed [Tutorial 01: Getting Started](./01-getting-started.md)
- Basic understanding of HTML and Go templates
- Familiarity with web development concepts

## 🚀 Getting Started

### Basic Template Setup

```go
package main

import (
    "html/template"
    "path/filepath"
    
    "github.com/go-zoox/zoox"
)

func main() {
    app := zoox.New()
    
    // Setup template directory
    app.SetTemplateDir("templates")
    
    // Basic template route
    app.Get("/", func(ctx *zoox.Context) {
        data := map[string]interface{}{
            "Title": "Welcome to Zoox",
            "Message": "Hello, World!",
            "User": map[string]string{
                "Name": "John Doe",
                "Email": "john@example.com",
            },
        }
        
        ctx.HTML(200, "index.html", data)
    })
    
    app.Listen(":8080")
}
```

### Template Inheritance System

```go
package main

import (
    "html/template"
    "path/filepath"
    "strings"
    
    "github.com/go-zoox/zoox"
)

type TemplateEngine struct {
    templates map[string]*template.Template
    funcMap   template.FuncMap
}

func NewTemplateEngine() *TemplateEngine {
    return &TemplateEngine{
        templates: make(map[string]*template.Template),
        funcMap: template.FuncMap{
            "upper": strings.ToUpper,
            "lower": strings.ToLower,
            "title": strings.Title,
            "join":  strings.Join,
        },
    }
}

func (te *TemplateEngine) LoadTemplates(dir string) error {
    layouts, err := filepath.Glob(filepath.Join(dir, "layouts", "*.html"))
    if err != nil {
        return err
    }
    
    pages, err := filepath.Glob(filepath.Join(dir, "pages", "*.html"))
    if err != nil {
        return err
    }
    
    for _, page := range pages {
        name := filepath.Base(page)
        files := append(layouts, page)
        
        tmpl, err := template.New(name).Funcs(te.funcMap).ParseFiles(files...)
        if err != nil {
            return err
        }
        
        te.templates[name] = tmpl
    }
    
    return nil
}

func (te *TemplateEngine) Render(name string, data interface{}) (string, error) {
    tmpl, exists := te.templates[name]
    if !exists {
        return "", fmt.Errorf("template %s not found", name)
    }
    
    var buf strings.Builder
    err := tmpl.Execute(&buf, data)
    return buf.String(), err
}

func main() {
    app := zoox.New()
    
    // Create template engine
    engine := NewTemplateEngine()
    if err := engine.LoadTemplates("templates"); err != nil {
        log.Fatal("Failed to load templates:", err)
    }
    
    // Use custom template engine
    app.Use(func(ctx *zoox.Context) {
        ctx.Set("templateEngine", engine)
        ctx.Next()
    })
    
    // Routes with template inheritance
    app.Get("/", func(ctx *zoox.Context) {
        engine := ctx.Get("templateEngine").(*TemplateEngine)
        
        data := map[string]interface{}{
            "Title": "Home Page",
            "Content": "Welcome to our website!",
            "Navigation": []map[string]string{
                {"Name": "Home", "URL": "/"},
                {"Name": "About", "URL": "/about"},
                {"Name": "Contact", "URL": "/contact"},
            },
        }
        
        html, err := engine.Render("home.html", data)
        if err != nil {
            ctx.JSON(500, map[string]string{"error": err.Error()})
            return
        }
        
        ctx.HTML(200, html, nil)
    })
    
    app.Listen(":8080")
}
```

### Advanced Template Features

```go
package main

import (
    "fmt"
    "html/template"
    "strings"
    "time"
    
    "github.com/go-zoox/zoox"
)

type AdvancedTemplateEngine struct {
    templates map[string]*template.Template
    funcMap   template.FuncMap
    cache     map[string]CachedTemplate
}

type CachedTemplate struct {
    Content   string
    ExpiresAt time.Time
}

func NewAdvancedTemplateEngine() *AdvancedTemplateEngine {
    return &AdvancedTemplateEngine{
        templates: make(map[string]*template.Template),
        cache:     make(map[string]CachedTemplate),
        funcMap: template.FuncMap{
            // String functions
            "upper":    strings.ToUpper,
            "lower":    strings.ToLower,
            "title":    strings.Title,
            "join":     strings.Join,
            "split":    strings.Split,
            "contains": strings.Contains,
            
            // Date functions
            "now":        time.Now,
            "formatDate": func(t time.Time, layout string) string {
                return t.Format(layout)
            },
            "timeAgo": func(t time.Time) string {
                duration := time.Since(t)
                switch {
                case duration < time.Minute:
                    return "just now"
                case duration < time.Hour:
                    return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
                case duration < 24*time.Hour:
                    return fmt.Sprintf("%d hours ago", int(duration.Hours()))
                default:
                    return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
                }
            },
            
            // Utility functions
            "add": func(a, b int) int { return a + b },
            "sub": func(a, b int) int { return a - b },
            "mul": func(a, b int) int { return a * b },
            "div": func(a, b int) int { return a / b },
            
            // Array functions
            "slice": func(items []interface{}, start, end int) []interface{} {
                if start < 0 || end > len(items) || start > end {
                    return []interface{}{}
                }
                return items[start:end]
            },
            "len": func(items interface{}) int {
                switch v := items.(type) {
                case []interface{}:
                    return len(v)
                case []string:
                    return len(v)
                case string:
                    return len(v)
                default:
                    return 0
                }
            },
            
            // Conditional functions
            "eq": func(a, b interface{}) bool { return a == b },
            "ne": func(a, b interface{}) bool { return a != b },
            "gt": func(a, b int) bool { return a > b },
            "lt": func(a, b int) bool { return a < b },
            "gte": func(a, b int) bool { return a >= b },
            "lte": func(a, b int) bool { return a <= b },
        },
    }
}

func main() {
    app := zoox.New()
    
    engine := NewAdvancedTemplateEngine()
    
    // Blog example with advanced templates
    app.Get("/blog", func(ctx *zoox.Context) {
        posts := []map[string]interface{}{
            {
                "Title":     "Getting Started with Zoox",
                "Content":   "Learn how to build web applications with Zoox framework...",
                "Author":    "John Doe",
                "CreatedAt": time.Now().Add(-2 * time.Hour),
                "Tags":      []string{"zoox", "golang", "web"},
                "Views":     125,
            },
            {
                "Title":     "Advanced Routing Techniques",
                "Content":   "Explore advanced routing patterns and best practices...",
                "Author":    "Jane Smith",
                "CreatedAt": time.Now().Add(-1 * 24 * time.Hour),
                "Tags":      []string{"routing", "advanced", "patterns"},
                "Views":     89,
            },
        }
        
        data := map[string]interface{}{
            "Title":       "Blog Posts",
            "Posts":       posts,
            "CurrentUser": "admin",
            "TotalPosts":  len(posts),
        }
        
        template := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>{{.Title}}</title>
            <style>
                body { font-family: Arial, sans-serif; margin: 20px; }
                .post { border: 1px solid #ddd; padding: 20px; margin: 20px 0; }
                .meta { color: #666; font-size: 0.9em; }
                .tags { margin-top: 10px; }
                .tag { background: #eee; padding: 2px 8px; margin: 2px; border-radius: 3px; }
            </style>
        </head>
        <body>
            <h1>{{.Title}} ({{.TotalPosts}} total)</h1>
            
            {{range .Posts}}
            <div class="post">
                <h2>{{.Title}}</h2>
                <div class="meta">
                    By {{.Author}} • {{timeAgo .CreatedAt}} • {{.Views}} views
                </div>
                <p>{{.Content}}</p>
                <div class="tags">
                    {{range .Tags}}
                    <span class="tag">{{.}}</span>
                    {{end}}
                </div>
            </div>
            {{end}}
            
            {{if eq .CurrentUser "admin"}}
            <p><a href="/admin/posts">Manage Posts</a></p>
            {{end}}
        </body>
        </html>
        `
        
        ctx.HTML(200, template, data)
    })
    
    app.Listen(":8080")
}
```

## 🎯 Hands-on Exercise

Create a complete blog system with template inheritance:

### Solution

```go
package main

import (
    "fmt"
    "html/template"
    "log"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/go-zoox/zoox"
)

type BlogSystem struct {
    engine *TemplateEngine
    posts  []Post
}

type Post struct {
    ID       int       `json:"id"`
    Title    string    `json:"title"`
    Content  string    `json:"content"`
    Author   string    `json:"author"`
    Date     time.Time `json:"date"`
    Tags     []string  `json:"tags"`
    Category string    `json:"category"`
}

func NewBlogSystem() *BlogSystem {
    return &BlogSystem{
        engine: NewTemplateEngine(),
        posts: []Post{
            {
                ID:       1,
                Title:    "Welcome to Our Blog",
                Content:  "This is our first blog post. Welcome to our journey!",
                Author:   "Admin",
                Date:     time.Now().Add(-24 * time.Hour),
                Tags:     []string{"welcome", "introduction"},
                Category: "General",
            },
            {
                ID:       2,
                Title:    "Learning Go Programming",
                Content:  "Go is a powerful programming language. Let's explore its features.",
                Author:   "John Doe",
                Date:     time.Now().Add(-12 * time.Hour),
                Tags:     []string{"go", "programming", "tutorial"},
                Category: "Technology",
            },
        },
    }
}

func (bs *BlogSystem) Setup(app *zoox.Application) {
    // Load templates
    if err := bs.engine.LoadTemplates("templates"); err != nil {
        log.Printf("Warning: Could not load templates: %v", err)
    }
    
    // Routes
    app.Get("/", bs.homePage)
    app.Get("/post/:id", bs.postPage)
    app.Get("/category/:category", bs.categoryPage)
    app.Get("/search", bs.searchPage)
}

func (bs *BlogSystem) homePage(ctx *zoox.Context) {
    data := map[string]interface{}{
        "Title":      "My Blog",
        "Posts":      bs.posts,
        "Categories": bs.getCategories(),
        "RecentPosts": bs.getRecentPosts(3),
    }
    
    bs.renderTemplate(ctx, "home.html", data)
}

func (bs *BlogSystem) postPage(ctx *zoox.Context) {
    id := ctx.Param("id")
    post := bs.getPostByID(id)
    
    if post == nil {
        ctx.JSON(404, map[string]string{"error": "Post not found"})
        return
    }
    
    data := map[string]interface{}{
        "Title":       post.Title,
        "Post":        post,
        "RelatedPosts": bs.getRelatedPosts(post, 3),
    }
    
    bs.renderTemplate(ctx, "post.html", data)
}

func (bs *BlogSystem) renderTemplate(ctx *zoox.Context, templateName string, data interface{}) {
    if html, err := bs.engine.Render(templateName, data); err == nil {
        ctx.HTML(200, html, nil)
    } else {
        // Fallback to inline template
        bs.renderInlineTemplate(ctx, templateName, data)
    }
}

func (bs *BlogSystem) renderInlineTemplate(ctx *zoox.Context, templateName string, data interface{}) {
    var template string
    
    switch templateName {
    case "home.html":
        template = `
        <!DOCTYPE html>
        <html>
        <head>
            <title>{{.Title}}</title>
            <style>
                body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
                .header { background: #333; color: white; padding: 20px; margin: -20px -20px 20px -20px; }
                .post { border: 1px solid #ddd; padding: 20px; margin: 20px 0; border-radius: 5px; }
                .meta { color: #666; font-size: 0.9em; margin-bottom: 10px; }
                .tags { margin-top: 10px; }
                .tag { background: #eee; padding: 2px 8px; margin: 2px; border-radius: 3px; font-size: 0.8em; }
                .sidebar { float: right; width: 300px; margin-left: 20px; }
                .main { margin-right: 320px; }
                .widget { background: #f9f9f9; padding: 15px; margin-bottom: 20px; border-radius: 5px; }
            </style>
        </head>
        <body>
            <div class="header">
                <h1>{{.Title}}</h1>
                <p>A simple blog built with Zoox</p>
            </div>
            
            <div class="sidebar">
                <div class="widget">
                    <h3>Categories</h3>
                    <ul>
                        {{range .Categories}}
                        <li><a href="/category/{{.}}">{{.}}</a></li>
                        {{end}}
                    </ul>
                </div>
                
                <div class="widget">
                    <h3>Recent Posts</h3>
                    <ul>
                        {{range .RecentPosts}}
                        <li><a href="/post/{{.ID}}">{{.Title}}</a></li>
                        {{end}}
                    </ul>
                </div>
            </div>
            
            <div class="main">
                <h2>Latest Posts</h2>
                {{range .Posts}}
                <div class="post">
                    <h3><a href="/post/{{.ID}}">{{.Title}}</a></h3>
                    <div class="meta">
                        By {{.Author}} on {{formatDate .Date "January 2, 2006"}} in {{.Category}}
                    </div>
                    <p>{{.Content}}</p>
                    <div class="tags">
                        {{range .Tags}}
                        <span class="tag">{{.}}</span>
                        {{end}}
                    </div>
                </div>
                {{end}}
            </div>
        </body>
        </html>
        `
    case "post.html":
        template = `
        <!DOCTYPE html>
        <html>
        <head>
            <title>{{.Title}}</title>
            <style>
                body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
                .header { background: #333; color: white; padding: 20px; margin: -20px -20px 20px -20px; }
                .post { padding: 20px; }
                .meta { color: #666; font-size: 0.9em; margin-bottom: 20px; }
                .content { line-height: 1.6; margin-bottom: 20px; }
                .tags { margin: 20px 0; }
                .tag { background: #eee; padding: 2px 8px; margin: 2px; border-radius: 3px; font-size: 0.8em; }
                .related { margin-top: 40px; padding-top: 20px; border-top: 1px solid #ddd; }
                .back { margin-bottom: 20px; }
            </style>
        </head>
        <body>
            <div class="header">
                <h1>{{.Post.Title}}</h1>
            </div>
            
            <div class="post">
                <div class="back">
                    <a href="/">&larr; Back to Home</a>
                </div>
                
                <div class="meta">
                    By {{.Post.Author}} on {{formatDate .Post.Date "January 2, 2006"}} in {{.Post.Category}}
                </div>
                
                <div class="content">
                    {{.Post.Content}}
                </div>
                
                <div class="tags">
                    {{range .Post.Tags}}
                    <span class="tag">{{.}}</span>
                    {{end}}
                </div>
                
                {{if .RelatedPosts}}
                <div class="related">
                    <h3>Related Posts</h3>
                    <ul>
                        {{range .RelatedPosts}}
                        <li><a href="/post/{{.ID}}">{{.Title}}</a></li>
                        {{end}}
                    </ul>
                </div>
                {{end}}
            </div>
        </body>
        </html>
        `
    }
    
    ctx.HTML(200, template, data)
}

func (bs *BlogSystem) getPostByID(id string) *Post {
    for i, post := range bs.posts {
        if fmt.Sprintf("%d", post.ID) == id {
            return &bs.posts[i]
        }
    }
    return nil
}

func (bs *BlogSystem) getCategories() []string {
    categories := make(map[string]bool)
    for _, post := range bs.posts {
        categories[post.Category] = true
    }
    
    result := make([]string, 0, len(categories))
    for category := range categories {
        result = append(result, category)
    }
    return result
}

func (bs *BlogSystem) getRecentPosts(limit int) []Post {
    if len(bs.posts) <= limit {
        return bs.posts
    }
    return bs.posts[:limit]
}

func (bs *BlogSystem) getRelatedPosts(post *Post, limit int) []Post {
    var related []Post
    for _, p := range bs.posts {
        if p.ID != post.ID && p.Category == post.Category {
            related = append(related, p)
            if len(related) >= limit {
                break
            }
        }
    }
    return related
}

func main() {
    app := zoox.New()
    
    blog := NewBlogSystem()
    blog.Setup(app)
    
    log.Println("Blog server starting on :8080")
    log.Println("Visit: http://localhost:8080")
    
    app.Listen(":8080")
}
```

## 📚 Key Takeaways

1. **Template Organization**: Use layouts and inheritance for maintainable templates
2. **Custom Functions**: Extend templates with custom helper functions
3. **Data Binding**: Efficiently pass data from handlers to templates
4. **Performance**: Cache templates and optimize rendering
5. **Error Handling**: Provide fallbacks for missing templates

## 🎯 Next Steps

- Learn [Tutorial 07: Static Files & Assets](./07-static-files-assets.md)
- Explore [Tutorial 08: WebSocket Development](./08-websocket-development.md)
- Study [Tutorial 10: Authentication & Authorization](./10-authentication-authorization.md)

---

**Congratulations!** You've mastered template engines in Zoox and can now build dynamic web applications with powerful templating capabilities. 