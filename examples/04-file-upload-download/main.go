package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-zoox/zoox"
	"github.com/go-zoox/zoox/middleware"
)

const (
	uploadDir = "./uploads"
	maxFileSize = 10 * 1024 * 1024 // 10MB
)

func main() {
	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Failed to create upload directory:", err)
	}

	app := zoox.Default()

	// Enable CORS for file uploads
	app.Use(middleware.CORS())

	// Body size limit for uploads
	app.Use(middleware.BodyLimit(&middleware.BodyLimitConfig{
		MaxSize: maxFileSize,
	}))

	// Serve static files
	app.Static("/uploads", uploadDir)

	// Main page with upload form
	app.Get("/", func(ctx *zoox.Context) {
		ctx.HTML(200, uploadPageHTML)
	})

	// File upload endpoints
	app.Post("/upload/single", uploadSingleFileHandler)
	app.Post("/upload/multiple", uploadMultipleFilesHandler)
	app.Post("/upload/chunked", uploadChunkedFileHandler)

	// File management endpoints
	app.Get("/api/files", listFilesHandler)
	app.Get("/api/files/:filename", getFileInfoHandler)
	app.Delete("/api/files/:filename", deleteFileHandler)

	// Download endpoints
	app.Get("/download/:filename", downloadFileHandler)
	app.Get("/download/:filename/inline", viewFileHandler)

	// Image processing endpoints
	app.Get("/image/:filename/thumbnail", thumbnailHandler)
	app.Get("/image/:filename/resize", resizeHandler)

	// File validation example
	app.Post("/upload/images", uploadImagesOnlyHandler)
	app.Post("/upload/documents", uploadDocumentsOnlyHandler)

	log.Println("File Upload/Download Server starting on http://localhost:8080")
	log.Println("Upload directory:", uploadDir)
	log.Println("Max file size:", maxFileSize/(1024*1024), "MB")
	log.Println("\nEndpoints:")
	log.Println("  GET  / - Upload form")
	log.Println("  POST /upload/single - Single file upload")
	log.Println("  POST /upload/multiple - Multiple files upload")
	log.Println("  GET  /api/files - List uploaded files")
	log.Println("  GET  /download/:filename - Download file")
	
	app.Run(":8080")
}

func uploadSingleFileHandler(ctx *zoox.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.JSON(400, zoox.H{
			"error": "No file uploaded",
			"details": err.Error(),
		})
		return
	}

	// Validate file
	if err := validateFile(file); err != nil {
		ctx.JSON(400, zoox.H{
			"error": "File validation failed",
			"details": err.Error(),
		})
		return
	}

	// Generate unique filename
	filename := generateUniqueFilename(file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	// Save file
	if err := ctx.SaveFile(file, filepath); err != nil {
		ctx.JSON(500, zoox.H{
			"error": "Failed to save file",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(200, zoox.H{
		"message": "File uploaded successfully",
		"file": zoox.H{
			"original_name": file.Filename,
			"saved_name":    filename,
			"size":          file.Size,
			"type":          file.Header.Get("Content-Type"),
			"url":           "/uploads/" + filename,
			"download_url":  "/download/" + filename,
		},
	})
}

func uploadMultipleFilesHandler(ctx *zoox.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(400, zoox.H{
			"error": "Failed to parse multipart form",
			"details": err.Error(),
		})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(400, zoox.H{
			"error": "No files uploaded",
		})
		return
	}

	var uploadedFiles []zoox.H
	var errors []string

	for _, file := range files {
		// Validate file
		if err := validateFile(file); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
			continue
		}

		// Generate unique filename
		filename := generateUniqueFilename(file.Filename)
		filepath := filepath.Join(uploadDir, filename)

		// Save file
		if err := ctx.SaveFile(file, filepath); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
			continue
		}

		uploadedFiles = append(uploadedFiles, zoox.H{
			"original_name": file.Filename,
			"saved_name":    filename,
			"size":          file.Size,
			"type":          file.Header.Get("Content-Type"),
			"url":           "/uploads/" + filename,
			"download_url":  "/download/" + filename,
		})
	}

	response := zoox.H{
		"message": fmt.Sprintf("Uploaded %d files successfully", len(uploadedFiles)),
		"files":   uploadedFiles,
		"total":   len(uploadedFiles),
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["failed"] = len(errors)
	}

	ctx.JSON(200, response)
}

func uploadChunkedFileHandler(ctx *zoox.Context) {
	// Get chunk information
	chunkNumber := ctx.Form().Get("chunkNumber", "0")
	totalChunks := ctx.Form().Get("totalChunks", "1")
	filename := ctx.Form().Get("filename")

	if filename == "" {
		ctx.JSON(400, zoox.H{"error": "Filename is required"})
		return
	}

	file, err := ctx.FormFile("chunk")
	if err != nil {
		ctx.JSON(400, zoox.H{
			"error": "No chunk uploaded",
			"details": err.Error(),
		})
		return
	}

	// Create temporary directory for chunks
	tempDir := filepath.Join(uploadDir, "temp", filename)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		ctx.JSON(500, zoox.H{
			"error": "Failed to create temp directory",
			"details": err.Error(),
		})
		return
	}

	// Save chunk
	chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%s", chunkNumber))
	if err := ctx.SaveFile(file, chunkPath); err != nil {
		ctx.JSON(500, zoox.H{
			"error": "Failed to save chunk",
			"details": err.Error(),
		})
		return
	}

	// Check if all chunks are uploaded
	totalChunksInt, _ := strconv.Atoi(totalChunks)
	chunkNumberInt, _ := strconv.Atoi(chunkNumber)

	if chunkNumberInt+1 == totalChunksInt {
		// Merge chunks
		finalPath := filepath.Join(uploadDir, generateUniqueFilename(filename))
		if err := mergeChunks(tempDir, finalPath, totalChunksInt); err != nil {
			ctx.JSON(500, zoox.H{
				"error": "Failed to merge chunks",
				"details": err.Error(),
			})
			return
		}

		// Clean up temp directory
		os.RemoveAll(tempDir)

		ctx.JSON(200, zoox.H{
			"message": "File uploaded successfully",
			"file": zoox.H{
				"original_name": filename,
				"saved_name":    filepath.Base(finalPath),
				"url":           "/uploads/" + filepath.Base(finalPath),
				"download_url":  "/download/" + filepath.Base(finalPath),
			},
		})
	} else {
		ctx.JSON(200, zoox.H{
			"message": fmt.Sprintf("Chunk %d uploaded successfully", chunkNumberInt+1),
			"chunk":   chunkNumberInt + 1,
			"total":   totalChunksInt,
		})
	}
}

func listFilesHandler(ctx *zoox.Context) {
	files, err := os.ReadDir(uploadDir)
	if err != nil {
		ctx.JSON(500, zoox.H{
			"error": "Failed to read upload directory",
			"details": err.Error(),
		})
		return
	}

	var fileList []zoox.H
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		fileList = append(fileList, zoox.H{
			"name":         file.Name(),
			"size":         info.Size(),
			"modified":     info.ModTime().Format(time.RFC3339),
			"url":          "/uploads/" + file.Name(),
			"download_url": "/download/" + file.Name(),
		})
	}

	ctx.JSON(200, zoox.H{
		"files": fileList,
		"total": len(fileList),
	})
}

func getFileInfoHandler(ctx *zoox.Context) {
	filename := ctx.Param().Get("filename")
	filepath := filepath.Join(uploadDir, filename)

	info, err := os.Stat(filepath)
	if err != nil {
		ctx.JSON(404, zoox.H{
			"error": "File not found",
		})
		return
	}

	ctx.JSON(200, zoox.H{
		"name":         filename,
		"size":         info.Size(),
		"modified":     info.ModTime().Format(time.RFC3339),
		"url":          "/uploads/" + filename,
		"download_url": "/download/" + filename,
	})
}

func deleteFileHandler(ctx *zoox.Context) {
	filename := ctx.Param().Get("filename")
	filepath := filepath.Join(uploadDir, filename)

	if err := os.Remove(filepath); err != nil {
		ctx.JSON(404, zoox.H{
			"error": "File not found or cannot be deleted",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(200, zoox.H{
		"message": "File deleted successfully",
		"filename": filename,
	})
}

func downloadFileHandler(ctx *zoox.Context) {
	filename := ctx.Param().Get("filename")
	filepath := filepath.Join(uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); err != nil {
		ctx.JSON(404, zoox.H{
			"error": "File not found",
		})
		return
	}

	// Force download
	ctx.Attachment(filepath, filename)
}

func viewFileHandler(ctx *zoox.Context) {
	filename := ctx.Param().Get("filename")
	filepath := filepath.Join(uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(filepath); err != nil {
		ctx.JSON(404, zoox.H{
			"error": "File not found",
		})
		return
	}

	// Serve file inline
	ctx.File(filepath)
}

func thumbnailHandler(ctx *zoox.Context) {
	filename := ctx.Param().Get("filename")
	
	// In a real application, you would generate thumbnails here
	ctx.JSON(200, zoox.H{
		"message": "Thumbnail generation not implemented",
		"filename": filename,
		"note": "This would generate a thumbnail of the image",
	})
}

func resizeHandler(ctx *zoox.Context) {
	filename := ctx.Param().Get("filename")
	width := ctx.Query().Get("width", "200")
	height := ctx.Query().Get("height", "200")
	
	// In a real application, you would resize images here
	ctx.JSON(200, zoox.H{
		"message": "Image resizing not implemented",
		"filename": filename,
		"width": width,
		"height": height,
		"note": "This would resize the image to specified dimensions",
	})
}

func uploadImagesOnlyHandler(ctx *zoox.Context) {
	file, err := ctx.FormFile("image")
	if err != nil {
		ctx.JSON(400, zoox.H{
			"error": "No image uploaded",
			"details": err.Error(),
		})
		return
	}

	// Validate image file
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		ctx.JSON(400, zoox.H{
			"error": "Only image files are allowed",
			"received": contentType,
		})
		return
	}

	filename := generateUniqueFilename(file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	if err := ctx.SaveFile(file, filepath); err != nil {
		ctx.JSON(500, zoox.H{
			"error": "Failed to save image",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(200, zoox.H{
		"message": "Image uploaded successfully",
		"image": zoox.H{
			"original_name": file.Filename,
			"saved_name":    filename,
			"size":          file.Size,
			"type":          contentType,
			"url":           "/uploads/" + filename,
			"thumbnail_url": "/image/" + filename + "/thumbnail",
		},
	})
}

func uploadDocumentsOnlyHandler(ctx *zoox.Context) {
	file, err := ctx.FormFile("document")
	if err != nil {
		ctx.JSON(400, zoox.H{
			"error": "No document uploaded",
			"details": err.Error(),
		})
		return
	}

	// Validate document file
	allowedTypes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"text/plain",
	}

	contentType := file.Header.Get("Content-Type")
	isAllowed := false
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		ctx.JSON(400, zoox.H{
			"error": "Only document files are allowed",
			"allowed": allowedTypes,
			"received": contentType,
		})
		return
	}

	filename := generateUniqueFilename(file.Filename)
	filepath := filepath.Join(uploadDir, filename)

	if err := ctx.SaveFile(file, filepath); err != nil {
		ctx.JSON(500, zoox.H{
			"error": "Failed to save document",
			"details": err.Error(),
		})
		return
	}

	ctx.JSON(200, zoox.H{
		"message": "Document uploaded successfully",
		"document": zoox.H{
			"original_name": file.Filename,
			"saved_name":    filename,
			"size":          file.Size,
			"type":          contentType,
			"url":           "/uploads/" + filename,
		},
	})
}

// Utility functions
func validateFile(file *multipart.FileHeader) error {
	if file.Size > maxFileSize {
		return fmt.Errorf("file size exceeds limit of %d MB", maxFileSize/(1024*1024))
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	dangerousExts := []string{".exe", ".bat", ".cmd", ".sh", ".ps1"}
	for _, dangerousExt := range dangerousExts {
		if ext == dangerousExt {
			return fmt.Errorf("file type %s is not allowed", ext)
		}
	}

	return nil
}

func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d%s", name, timestamp, ext)
}

func mergeChunks(tempDir, finalPath string, totalChunks int) error {
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return err
	}
	defer finalFile.Close()

	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(tempDir, fmt.Sprintf("chunk_%d", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(finalFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

const uploadPageHTML = `<!DOCTYPE html>
<html>
<head>
    <title>File Upload & Download</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        .upload-section { background: #f8f9fa; padding: 20px; border-radius: 5px; margin: 20px 0; }
        .upload-section h3 { margin-top: 0; }
        input[type="file"] { margin: 10px 0; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; margin: 5px; }
        button:hover { background: #0056b3; }
        .file-list { margin: 20px 0; }
        .file-item { background: #fff; padding: 10px; border: 1px solid #ddd; margin: 5px 0; border-radius: 3px; }
        .progress { background: #e9ecef; border-radius: 3px; height: 20px; margin: 10px 0; }
        .progress-bar { background: #007bff; height: 100%; border-radius: 3px; width: 0%; transition: width 0.3s; }
    </style>
</head>
<body>
    <h1>🚀 Zoox File Upload & Download</h1>

    <div class="upload-section">
        <h3>Single File Upload</h3>
        <input type="file" id="singleFile">
        <button onclick="uploadSingle()">Upload</button>
        <div id="singleProgress" class="progress" style="display: none;">
            <div class="progress-bar"></div>
        </div>
        <div id="singleResult"></div>
    </div>

    <div class="upload-section">
        <h3>Multiple Files Upload</h3>
        <input type="file" id="multipleFiles" multiple>
        <button onclick="uploadMultiple()">Upload All</button>
        <div id="multipleProgress" class="progress" style="display: none;">
            <div class="progress-bar"></div>
        </div>
        <div id="multipleResult"></div>
    </div>

    <div class="upload-section">
        <h3>Images Only</h3>
        <input type="file" id="imageFile" accept="image/*">
        <button onclick="uploadImage()">Upload Image</button>
        <div id="imageResult"></div>
    </div>

    <div class="upload-section">
        <h3>Documents Only</h3>
        <input type="file" id="documentFile" accept=".pdf,.doc,.docx,.txt">
        <button onclick="uploadDocument()">Upload Document</button>
        <div id="documentResult"></div>
    </div>

    <div class="file-list">
        <h3>Uploaded Files</h3>
        <button onclick="loadFiles()">Refresh List</button>
        <div id="filesList"></div>
    </div>

    <script>
        function uploadSingle() {
            const fileInput = document.getElementById('singleFile');
            const file = fileInput.files[0];
            if (!file) {
                alert('Please select a file');
                return;
            }

            const formData = new FormData();
            formData.append('file', file);

            uploadFile('/upload/single', formData, 'singleProgress', 'singleResult');
        }

        function uploadMultiple() {
            const fileInput = document.getElementById('multipleFiles');
            const files = fileInput.files;
            if (files.length === 0) {
                alert('Please select files');
                return;
            }

            const formData = new FormData();
            for (let i = 0; i < files.length; i++) {
                formData.append('files', files[i]);
            }

            uploadFile('/upload/multiple', formData, 'multipleProgress', 'multipleResult');
        }

        function uploadImage() {
            const fileInput = document.getElementById('imageFile');
            const file = fileInput.files[0];
            if (!file) {
                alert('Please select an image');
                return;
            }

            const formData = new FormData();
            formData.append('image', file);

            uploadFile('/upload/images', formData, null, 'imageResult');
        }

        function uploadDocument() {
            const fileInput = document.getElementById('documentFile');
            const file = fileInput.files[0];
            if (!file) {
                alert('Please select a document');
                return;
            }

            const formData = new FormData();
            formData.append('document', file);

            uploadFile('/upload/documents', formData, null, 'documentResult');
        }

        function uploadFile(url, formData, progressId, resultId) {
            const xhr = new XMLHttpRequest();
            
            if (progressId) {
                const progressDiv = document.getElementById(progressId);
                const progressBar = progressDiv.querySelector('.progress-bar');
                progressDiv.style.display = 'block';
                
                xhr.upload.addEventListener('progress', function(e) {
                    if (e.lengthComputable) {
                        const percentComplete = (e.loaded / e.total) * 100;
                        progressBar.style.width = percentComplete + '%';
                    }
                });
            }

            xhr.onload = function() {
                if (progressId) {
                    document.getElementById(progressId).style.display = 'none';
                }
                
                const result = JSON.parse(xhr.responseText);
                document.getElementById(resultId).innerHTML = 
                    '<pre>' + JSON.stringify(result, null, 2) + '</pre>';
                
                if (xhr.status === 200) {
                    loadFiles();
                }
            };

            xhr.onerror = function() {
                if (progressId) {
                    document.getElementById(progressId).style.display = 'none';
                }
                document.getElementById(resultId).innerHTML = 
                    '<div style="color: red;">Upload failed</div>';
            };

            xhr.open('POST', url);
            xhr.send(formData);
        }

        function loadFiles() {
            fetch('/api/files')
                .then(response => response.json())
                .then(data => {
                    const filesList = document.getElementById('filesList');
                    if (data.files && data.files.length > 0) {
                        filesList.innerHTML = data.files.map(file => 
                            '<div class="file-item">' +
                                '<strong>' + file.name + '</strong> ' +
                                '(' + formatFileSize(file.size) + ') ' +
                                '<a href="' + file.download_url + '" target="_blank">Download</a> ' +
                                '<button onclick="deleteFile(\'' + file.name + '\')">Delete</button>' +
                            '</div>'
                        ).join('');
                    } else {
                        filesList.innerHTML = '<p>No files uploaded yet.</p>';
                    }
                })
                .catch(error => {
                    console.error('Error loading files:', error);
                });
        }

        function deleteFile(filename) {
            if (!confirm('Are you sure you want to delete ' + filename + '?')) {
                return;
            }

            fetch('/api/files/' + encodeURIComponent(filename), {
                method: 'DELETE'
            })
            .then(response => response.json())
            .then(data => {
                alert(data.message || 'File deleted');
                loadFiles();
            })
            .catch(error => {
                console.error('Error deleting file:', error);
                alert('Failed to delete file');
            });
        }

        function formatFileSize(bytes) {
            if (bytes === 0) return '0 Bytes';
            const k = 1024;
            const sizes = ['Bytes', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        // Load files on page load
        window.onload = function() {
            loadFiles();
        };
    </script>
</body>
</html>` 