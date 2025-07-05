# Tutorial 07: Static Files & Assets

## 📖 Overview

Learn how to serve static files and optimize assets in Zoox applications. This tutorial covers static file serving, asset optimization, caching strategies, and CDN integration.

## 🎯 Learning Objectives

- Serve static files efficiently
- Implement asset optimization
- Configure caching strategies
- Integrate with CDN services
- Handle file uploads and downloads

## 📋 Prerequisites

- Completed [Tutorial 01: Getting Started](./01-getting-started.md)
- Basic understanding of web assets (CSS, JS, images)
- Familiarity with HTTP caching

## 🚀 Getting Started

### Basic Static File Serving

```go
package main

import (
    "github.com/go-zoox/zoox"
)

func main() {
    app := zoox.New()
    
    // Serve static files from public directory
    app.Static("/static", "./public")
    
    // Serve specific file types
    app.StaticFile("/favicon.ico", "./public/favicon.ico")
    
    // Custom static file handler
    app.Get("/assets/*", func(ctx *zoox.Context) {
        file := ctx.Param("*")
        ctx.File("./assets/" + file)
    })
    
    app.Listen(":8080")
}
```

### Advanced Asset Management

```go
package main

import (
    "crypto/md5"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/go-zoox/zoox"
)

type AssetManager struct {
    publicDir   string
    cacheMaxAge int
    hashes      map[string]string
}

func NewAssetManager(publicDir string) *AssetManager {
    return &AssetManager{
        publicDir:   publicDir,
        cacheMaxAge: 3600, // 1 hour
        hashes:      make(map[string]string),
    }
}

func (am *AssetManager) Setup(app *zoox.Application) {
    // Generate asset hashes for cache busting
    am.generateHashes()
    
    // Static file middleware with optimization
    app.Use(am.staticMiddleware())
    
    // Asset helper endpoint
    app.Get("/assets/manifest", func(ctx *zoox.Context) {
        ctx.JSON(200, map[string]interface{}{
            "hashes": am.hashes,
            "version": time.Now().Unix(),
        })
    })
}

func (am *AssetManager) staticMiddleware() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        path := ctx.Request.URL.Path
        
        // Check if it's a static asset request
        if strings.HasPrefix(path, "/assets/") {
            am.serveAsset(ctx, path)
            return
        }
        
        ctx.Next()
    }
}

func (am *AssetManager) serveAsset(ctx *zoox.Context, path string) {
    // Remove /assets/ prefix
    assetPath := strings.TrimPrefix(path, "/assets/")
    fullPath := filepath.Join(am.publicDir, assetPath)
    
    // Check if file exists
    if _, err := os.Stat(fullPath); os.IsNotExist(err) {
        ctx.Status(404)
        return
    }
    
    // Set cache headers
    ctx.Header("Cache-Control", fmt.Sprintf("max-age=%d", am.cacheMaxAge))
    ctx.Header("ETag", am.getFileHash(fullPath))
    
    // Check if client has cached version
    if ctx.Header("If-None-Match") == am.getFileHash(fullPath) {
        ctx.Status(304)
        return
    }
    
    // Set content type based on file extension
    ext := filepath.Ext(assetPath)
    switch ext {
    case ".css":
        ctx.Header("Content-Type", "text/css")
    case ".js":
        ctx.Header("Content-Type", "application/javascript")
    case ".png":
        ctx.Header("Content-Type", "image/png")
    case ".jpg", ".jpeg":
        ctx.Header("Content-Type", "image/jpeg")
    case ".gif":
        ctx.Header("Content-Type", "image/gif")
    case ".svg":
        ctx.Header("Content-Type", "image/svg+xml")
    }
    
    // Serve the file
    ctx.File(fullPath)
}

func (am *AssetManager) generateHashes() {
    filepath.Walk(am.publicDir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() {
            return err
        }
        
        relPath, _ := filepath.Rel(am.publicDir, path)
        am.hashes[relPath] = am.getFileHash(path)
        return nil
    })
}

func (am *AssetManager) getFileHash(path string) string {
    if hash, exists := am.hashes[path]; exists {
        return hash
    }
    
    file, err := os.Open(path)
    if err != nil {
        return ""
    }
    defer file.Close()
    
    hasher := md5.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return ""
    }
    
    return fmt.Sprintf("%x", hasher.Sum(nil))
}

func main() {
    app := zoox.New()
    
    // Setup asset manager
    assets := NewAssetManager("./public")
    assets.Setup(app)
    
    // Main page with assets
    app.Get("/", func(ctx *zoox.Context) {
        html := `
        <!DOCTYPE html>
        <html>
        <head>
            <title>Asset Management Demo</title>
            <link rel="stylesheet" href="/assets/css/styles.css">
        </head>
        <body>
            <h1>Welcome to Asset Management</h1>
            <p>This page demonstrates optimized asset serving.</p>
            <script src="/assets/js/app.js"></script>
        </body>
        </html>
        `
        ctx.HTML(200, html, nil)
    })
    
    app.Listen(":8080")
}
```

### File Upload System

```go
package main

import (
    "crypto/rand"
    "fmt"
    "io"
    "mime/multipart"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/go-zoox/zoox"
)

type FileUploadManager struct {
    uploadDir   string
    maxFileSize int64
    allowedExts []string
}

func NewFileUploadManager(uploadDir string) *FileUploadManager {
    return &FileUploadManager{
        uploadDir:   uploadDir,
        maxFileSize: 10 << 20, // 10MB
        allowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"},
    }
}

func (fum *FileUploadManager) Setup(app *zoox.Application) {
    // Ensure upload directory exists
    os.MkdirAll(fum.uploadDir, 0755)
    
    // Upload endpoints
    app.Post("/upload", fum.handleUpload)
    app.Get("/uploads/*", fum.serveUpload)
    app.Delete("/uploads/:filename", fum.deleteUpload)
    
    // Upload form
    app.Get("/upload-form", fum.uploadForm)
}

func (fum *FileUploadManager) handleUpload(ctx *zoox.Context) {
    // Parse multipart form
    err := ctx.Request.ParseMultipartForm(fum.maxFileSize)
    if err != nil {
        ctx.JSON(400, map[string]string{"error": "File too large"})
        return
    }
    
    file, header, err := ctx.Request.FormFile("file")
    if err != nil {
        ctx.JSON(400, map[string]string{"error": "No file uploaded"})
        return
    }
    defer file.Close()
    
    // Validate file
    if !fum.isAllowedFile(header.Filename) {
        ctx.JSON(400, map[string]string{"error": "File type not allowed"})
        return
    }
    
    // Generate unique filename
    filename := fum.generateUniqueFilename(header.Filename)
    filepath := filepath.Join(fum.uploadDir, filename)
    
    // Save file
    if err := fum.saveFile(file, filepath); err != nil {
        ctx.JSON(500, map[string]string{"error": "Failed to save file"})
        return
    }
    
    ctx.JSON(200, map[string]interface{}{
        "message":  "File uploaded successfully",
        "filename": filename,
        "url":      "/uploads/" + filename,
        "size":     header.Size,
    })
}

func (fum *FileUploadManager) serveUpload(ctx *zoox.Context) {
    filename := ctx.Param("*")
    filepath := filepath.Join(fum.uploadDir, filename)
    
    // Security check - prevent directory traversal
    if strings.Contains(filename, "..") {
        ctx.Status(403)
        return
    }
    
    // Check if file exists
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        ctx.Status(404)
        return
    }
    
    ctx.File(filepath)
}

func (fum *FileUploadManager) deleteUpload(ctx *zoox.Context) {
    filename := ctx.Param("filename")
    filepath := filepath.Join(fum.uploadDir, filename)
    
    // Security check
    if strings.Contains(filename, "..") {
        ctx.JSON(403, map[string]string{"error": "Invalid filename"})
        return
    }
    
    if err := os.Remove(filepath); err != nil {
        ctx.JSON(500, map[string]string{"error": "Failed to delete file"})
        return
    }
    
    ctx.JSON(200, map[string]string{"message": "File deleted successfully"})
}

func (fum *FileUploadManager) uploadForm(ctx *zoox.Context) {
    html := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>File Upload</title>
        <style>
            body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
            .upload-area { border: 2px dashed #ccc; padding: 40px; text-align: center; margin: 20px 0; }
            .upload-area.dragover { border-color: #007bff; background-color: #f8f9fa; }
            input[type="file"] { margin: 20px 0; }
            button { background: #007bff; color: white; padding: 10px 20px; border: none; cursor: pointer; }
            .result { margin-top: 20px; padding: 10px; background: #f8f9fa; border-radius: 5px; }
        </style>
    </head>
    <body>
        <h1>File Upload Demo</h1>
        
        <div class="upload-area" id="uploadArea">
            <p>Drag and drop files here or click to select</p>
            <input type="file" id="fileInput" multiple>
            <button onclick="uploadFiles()">Upload Files</button>
        </div>
        
        <div id="result" class="result" style="display: none;"></div>
        
        <script>
            const uploadArea = document.getElementById('uploadArea');
            const fileInput = document.getElementById('fileInput');
            const result = document.getElementById('result');
            
            uploadArea.addEventListener('dragover', (e) => {
                e.preventDefault();
                uploadArea.classList.add('dragover');
            });
            
            uploadArea.addEventListener('dragleave', () => {
                uploadArea.classList.remove('dragover');
            });
            
            uploadArea.addEventListener('drop', (e) => {
                e.preventDefault();
                uploadArea.classList.remove('dragover');
                fileInput.files = e.dataTransfer.files;
            });
            
            uploadArea.addEventListener('click', () => {
                fileInput.click();
            });
            
            async function uploadFiles() {
                const files = fileInput.files;
                if (files.length === 0) {
                    alert('Please select files to upload');
                    return;
                }
                
                const results = [];
                for (let file of files) {
                    const formData = new FormData();
                    formData.append('file', file);
                    
                    try {
                        const response = await fetch('/upload', {
                            method: 'POST',
                            body: formData
                        });
                        
                        const data = await response.json();
                        results.push(data);
                    } catch (error) {
                        results.push({ error: 'Upload failed for ' + file.name });
                    }
                }
                
                displayResults(results);
            }
            
            function displayResults(results) {
                result.style.display = 'block';
                result.innerHTML = '<h3>Upload Results:</h3>';
                
                results.forEach(item => {
                    const div = document.createElement('div');
                    if (item.error) {
                        div.innerHTML = '<p style="color: red;">' + item.error + '</p>';
                    } else {
                        div.innerHTML = '<p style="color: green;">✓ ' + item.filename + ' uploaded successfully</p>';
                    }
                    result.appendChild(div);
                });
            }
        </script>
    </body>
    </html>
    `
    
    ctx.HTML(200, html, nil)
}

func (fum *FileUploadManager) isAllowedFile(filename string) bool {
    ext := strings.ToLower(filepath.Ext(filename))
    for _, allowed := range fum.allowedExts {
        if ext == allowed {
            return true
        }
    }
    return false
}

func (fum *FileUploadManager) generateUniqueFilename(original string) string {
    ext := filepath.Ext(original)
    name := strings.TrimSuffix(original, ext)
    
    // Generate random suffix
    b := make([]byte, 8)
    rand.Read(b)
    suffix := fmt.Sprintf("%x", b)
    
    return fmt.Sprintf("%s_%d_%s%s", name, time.Now().Unix(), suffix, ext)
}

func (fum *FileUploadManager) saveFile(src multipart.File, dst string) error {
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer out.Close()
    
    _, err = io.Copy(out, src)
    return err
}

func main() {
    app := zoox.New()
    
    // Setup file upload manager
    uploader := NewFileUploadManager("./uploads")
    uploader.Setup(app)
    
    // Setup asset manager
    assets := NewAssetManager("./public")
    assets.Setup(app)
    
    app.Listen(":8080")
}
```

## 🎯 Hands-on Exercise

Create a complete asset management system with:
1. Static file serving with optimization
2. File upload with validation
3. Image resizing and optimization
4. CDN integration simulation

## 📚 Key Takeaways

1. **Static Serving**: Efficiently serve static files with proper headers
2. **Asset Optimization**: Implement caching and compression
3. **File Uploads**: Handle file uploads securely with validation
4. **Performance**: Use ETags and cache headers for optimization
5. **Security**: Validate file types and prevent directory traversal

## 🎯 Next Steps

- Learn [Tutorial 08: WebSocket Development](./08-websocket-development.md)
- Explore [Tutorial 09: JSON-RPC Services](./09-json-rpc-services.md)
- Study [Tutorial 10: Authentication & Authorization](./10-authentication-authorization.md)

---

**Congratulations!** You've mastered static file serving and asset management in Zoox! 