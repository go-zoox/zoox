# File Upload/Download System Example

This example demonstrates comprehensive file handling capabilities with the Zoox framework, including file uploads, downloads, validation, security measures, and metadata management.

## Features

### File Upload
- **Single and multiple file uploads** with drag-and-drop support
- **Chunked upload** for large files with progress tracking
- **File validation** including type, size, and content validation
- **Secure file storage** with sanitized filenames
- **Upload progress** tracking and cancellation

### File Download
- **Secure file serving** with access controls
- **Range request support** for partial downloads and resumable downloads
- **Content-Type detection** and proper headers
- **Download speed limiting** and bandwidth control

### Security Features
- **File type validation** using MIME types and magic numbers
- **Filename sanitization** to prevent directory traversal
- **Size limits** configurable per file type
- **Virus scanning integration** (placeholder for production)
- **Access control** with authentication

### File Management
- **File metadata storage** including upload time, user, size
- **File organization** with automatic directory structure
- **Duplicate detection** using file hashing
- **Cleanup utilities** for orphaned files
- **Storage quotas** per user or globally

## Quick Start

1. **Run the application:**
   ```bash
   cd examples/04-file-upload-download
   go mod tidy
   go run main.go
   ```

2. **Access the file interface:**
   - Open browser to `http://localhost:8080`
   - Try uploading various file types
   - Download files to test the system

3. **Test with curl:**
   ```bash
   # Upload a file
   curl -X POST -F "file=@example.txt" http://localhost:8080/upload
   
   # Download a file  
   curl -O http://localhost:8080/download/example.txt
   
   # Get file info
   curl http://localhost:8080/files/example.txt/info
   ```

## API Endpoints

### Upload Endpoints
- **POST /upload** - Single file upload
- **POST /upload/multiple** - Multiple file upload  
- **POST /upload/chunked** - Chunked upload for large files
- **GET /upload** - Upload interface (HTML form)

### Download Endpoints
- **GET /download/:filename** - Download file by name
- **GET /files/:id** - Download file by ID
- **HEAD /download/:filename** - Get file headers without content

### File Management
- **GET /files** - List all files with pagination
- **GET /files/:filename/info** - Get file metadata
- **DELETE /files/:filename** - Delete file (requires auth)
- **POST /files/:filename/move** - Move/rename file

### System Endpoints
- **GET /storage/stats** - Storage statistics and quotas
- **GET /health** - System health check
- **POST /cleanup** - Clean orphaned files (admin only)

## File Validation Rules

### Supported File Types
```go
var allowedTypes = map[string][]string{
    "image": {".jpg", ".jpeg", ".png", ".gif", ".webp"},
    "document": {".pdf", ".doc", ".docx", ".txt", ".md"},
    "archive": {".zip", ".tar", ".gz", ".7z"},
    "video": {".mp4", ".avi", ".mov", ".webm"},
    "audio": {".mp3", ".wav", ".aac", ".ogg"},
}
```

### Size Limits
- **Images:** 5MB maximum
- **Documents:** 10MB maximum  
- **Archives:** 50MB maximum
- **Videos:** 100MB maximum
- **Audio:** 25MB maximum
- **Default:** 10MB maximum

### Security Validations
1. **MIME type checking** against file extension
2. **Magic number validation** to detect file type spoofing
3. **Filename sanitization** to prevent path traversal
4. **Content scanning** for malicious content (placeholder)

## Upload Process Flow

### 1. File Validation
```go
func validateFile(header *multipart.FileHeader) error {
    // Check file size
    if header.Size > maxFileSize {
        return ErrFileTooLarge
    }
    
    // Validate file extension
    ext := filepath.Ext(header.Filename)
    if !isAllowedExtension(ext) {
        return ErrInvalidFileType  
    }
    
    // Additional validations...
}
```

### 2. Secure Storage
```go
func saveFile(file multipart.File, filename string) error {
    // Generate secure filename
    safeFilename := sanitizeFilename(filename)
    
    // Create directory structure
    dir := generateStoragePath(safeFilename)
    
    // Save with atomic write
    return atomicWrite(filepath.Join(dir, safeFilename), file)
}
```

### 3. Metadata Storage
```go
type FileMetadata struct {
    ID           string    `json:"id"`
    Filename     string    `json:"filename"`
    OriginalName string    `json:"original_name"`
    Size         int64     `json:"size"`
    ContentType  string    `json:"content_type"`
    Hash         string    `json:"hash"`
    UploadedAt   time.Time `json:"uploaded_at"`
    UploadedBy   string    `json:"uploaded_by"`
}
```

## Download Features

### Range Request Support
```bash
# Download bytes 100-199
curl -H "Range: bytes=100-199" http://localhost:8080/download/large-file.zip

# Resume download from byte 1000
curl -H "Range: bytes=1000-" http://localhost:8080/download/large-file.zip
```

### Content Disposition
- **Inline viewing** for images and PDFs
- **Attachment download** for archives and executables  
- **Custom filenames** for downloaded files

### Streaming Downloads
- **Memory efficient** streaming for large files
- **Progress tracking** with content-length headers
- **Bandwidth limiting** to prevent server overload

## Testing Scenarios

### 1. Basic Upload/Download
```bash
# Create test file
echo "Hello World" > test.txt

# Upload file
curl -X POST -F "file=@test.txt" http://localhost:8080/upload

# Verify upload
curl http://localhost:8080/files

# Download file
curl -O http://localhost:8080/download/test.txt
```

### 2. Multiple File Upload
```bash
# Upload multiple files
curl -X POST \
  -F "files=@file1.txt" \
  -F "files=@file2.jpg" \
  -F "files=@file3.pdf" \
  http://localhost:8080/upload/multiple
```

### 3. Large File Handling
```bash
# Create large file
dd if=/dev/zero of=large.dat bs=1M count=50

# Upload with progress
curl -X POST -F "file=@large.dat" \
  --progress-bar http://localhost:8080/upload

# Test range downloads
curl -H "Range: bytes=0-1023" \
  http://localhost:8080/download/large.dat
```

### 4. Security Testing
```bash
# Try path traversal
curl -X POST -F "file=@test.txt" \
  -F "filename=../../../etc/passwd" \
  http://localhost:8080/upload

# Try oversized file
dd if=/dev/zero of=huge.dat bs=1M count=200
curl -X POST -F "file=@huge.dat" http://localhost:8080/upload

# Try invalid file type  
curl -X POST -F "file=@malware.exe" http://localhost:8080/upload
```

## Configuration

### Environment Variables
```bash
export UPLOAD_DIR="./uploads"
export MAX_FILE_SIZE="10485760"  # 10MB
export MAX_STORAGE_SIZE="1073741824"  # 1GB
export ENABLE_VIRUS_SCAN="false"
export CLEANUP_INTERVAL="24h"
```

### Storage Configuration
```go
type StorageConfig struct {
    UploadDir       string        `json:"upload_dir"`
    MaxFileSize     int64         `json:"max_file_size"`
    MaxStorageSize  int64         `json:"max_storage_size"`
    AllowedTypes    []string      `json:"allowed_types"`
    CleanupInterval time.Duration `json:"cleanup_interval"`
}
```

## Production Considerations

### 1. Storage Backend
- **Local filesystem** for development
- **AWS S3/MinIO** for production scalability  
- **Database metadata** for file tracking
- **CDN integration** for fast downloads

### 2. Security Enhancements
- **Virus scanning** with ClamAV or similar
- **Content analysis** for malicious content
- **Rate limiting** for upload endpoints
- **Authentication** and authorization

### 3. Performance Optimization
- **Asynchronous processing** for large uploads
- **Image resizing** and thumbnail generation
- **Compression** for compatible file types
- **Caching** for frequently accessed files

### 4. Monitoring and Logging
- **Upload/download metrics** and analytics
- **Error tracking** and alerting
- **Storage usage** monitoring
- **Performance metrics** and optimization

## Learning Objectives

After working with this example, you will understand:

1. **File Handling in Go**
   - Multipart form processing
   - File I/O operations and streaming
   - Memory-efficient file processing

2. **Security Best Practices**  
   - Input validation and sanitization
   - Path traversal prevention
   - Content type validation

3. **HTTP File Operations**
   - Range request handling
   - Proper content headers
   - Upload progress tracking

4. **Production File Systems**
   - Scalable storage architectures
   - Metadata management
   - Error handling and recovery

## Extending the Example

### Add Image Processing
```go
import "github.com/disintegration/imaging"

func resizeImage(src image.Image, width, height int) image.Image {
    return imaging.Resize(src, width, height, imaging.Lanczos)
}
```

### Add Cloud Storage
```go
import "github.com/aws/aws-sdk-go/service/s3"

func uploadToS3(file io.Reader, key string) error {
    _, err := s3.PutObject(&s3.PutObjectInput{
        Bucket: aws.String("my-bucket"),
        Key:    aws.String(key),
        Body:   file,
    })
    return err
}
```

### Add File Sharing
```go
type ShareLink struct {
    FileID    string    `json:"file_id"`
    Token     string    `json:"token"`  
    ExpiresAt time.Time `json:"expires_at"`
    Downloads int       `json:"downloads"`
    MaxDownloads int    `json:"max_downloads"`
}
```

## Troubleshooting

**Upload fails with "file too large":**
```bash
# Check server limits
curl http://localhost:8080/storage/stats

# Increase limits in configuration
export MAX_FILE_SIZE="52428800"  # 50MB
```

**Download returns 404:**
```bash
# Verify file exists
curl http://localhost:8080/files

# Check file permissions
ls -la uploads/
```

**Performance issues with large files:**
```bash
# Monitor memory usage
go tool pprof http://localhost:8080/debug/pprof/heap

# Enable streaming for large files
# Use chunked upload for files > 10MB
```

## Next Steps

- Explore the **Production API** example for authentication integration
- Check the **Middleware Showcase** for security and monitoring patterns
- Review cloud storage integration patterns for scalability 