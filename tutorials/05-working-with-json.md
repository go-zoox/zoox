# Tutorial 05: Working with JSON

## Overview
JSON (JavaScript Object Notation) is the most common data format for web APIs. In this tutorial, you'll learn how to effectively work with JSON in Zoox applications, including parsing, validation, serialization, and handling complex data structures.

## Learning Objectives
- Parse JSON from requests
- Validate JSON data
- Handle nested JSON structures
- Custom JSON serialization
- Error handling for JSON operations
- Performance optimization for JSON processing

## Prerequisites
- Complete Tutorial 04: Middleware Basics
- Basic understanding of Go structs and interfaces
- Familiarity with JSON format

## JSON Parsing Fundamentals

### Basic JSON Parsing

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "github.com/go-zoox/zoox"
)

// User represents a user in our system
type User struct {
    ID       int       `json:"id"`
    Name     string    `json:"name" validate:"required,min=2,max=50"`
    Email    string    `json:"email" validate:"required,email"`
    Age      int       `json:"age" validate:"min=0,max=150"`
    Active   bool      `json:"active"`
    Created  time.Time `json:"created"`
    Profile  Profile   `json:"profile"`
    Tags     []string  `json:"tags"`
}

// Profile represents user profile information
type Profile struct {
    Bio     string `json:"bio"`
    Website string `json:"website"`
    Avatar  string `json:"avatar"`
}

// CreateUserRequest represents the request for creating a user
type CreateUserRequest struct {
    Name    string   `json:"name" validate:"required"`
    Email   string   `json:"email" validate:"required,email"`
    Age     int      `json:"age" validate:"min=0"`
    Profile Profile  `json:"profile"`
    Tags    []string `json:"tags"`
}

// Response represents a standard API response
type Response struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func main() {
    app := zoox.New()

    // Middleware for JSON content type
    app.Use(func(ctx *zoox.Context) {
        ctx.Set("Content-Type", "application/json")
        ctx.Next()
    })

    // Basic JSON parsing
    app.Post("/users", createUser)
    
    // Complex JSON parsing
    app.Post("/users/batch", createBatchUsers)
    
    // JSON validation
    app.Put("/users/:id", updateUser)
    
    // Custom JSON serialization
    app.Get("/users/:id", getUser)
    
    // Handle JSON arrays
    app.Post("/users/search", searchUsers)

    fmt.Println("Server starting on :8080")
    log.Fatal(app.Listen(":8080"))
}

func createUser(ctx *zoox.Context) {
    var req CreateUserRequest
    
    // Parse JSON from request body
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Invalid JSON format",
            Error:   err.Error(),
        })
        return
    }
    
    // Validate required fields
    if req.Name == "" {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Name is required",
        })
        return
    }
    
    if req.Email == "" {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Email is required",
        })
        return
    }
    
    // Create user
    user := User{
        ID:      generateID(),
        Name:    req.Name,
        Email:   req.Email,
        Age:     req.Age,
        Active:  true,
        Created: time.Now(),
        Profile: req.Profile,
        Tags:    req.Tags,
    }
    
    ctx.JSON(http.StatusCreated, Response{
        Success: true,
        Message: "User created successfully",
        Data:    user,
    })
}

func createBatchUsers(ctx *zoox.Context) {
    var requests []CreateUserRequest
    
    // Parse JSON array
    if err := ctx.BindJSON(&requests); err != nil {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Invalid JSON array format",
            Error:   err.Error(),
        })
        return
    }
    
    if len(requests) == 0 {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "At least one user is required",
        })
        return
    }
    
    if len(requests) > 100 {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Maximum 100 users allowed per batch",
        })
        return
    }
    
    var users []User
    var errors []string
    
    for i, req := range requests {
        // Validate each user
        if req.Name == "" {
            errors = append(errors, fmt.Sprintf("User %d: Name is required", i+1))
            continue
        }
        
        if req.Email == "" {
            errors = append(errors, fmt.Sprintf("User %d: Email is required", i+1))
            continue
        }
        
        user := User{
            ID:      generateID(),
            Name:    req.Name,
            Email:   req.Email,
            Age:     req.Age,
            Active:  true,
            Created: time.Now(),
            Profile: req.Profile,
            Tags:    req.Tags,
        }
        
        users = append(users, user)
    }
    
    if len(errors) > 0 {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Validation errors occurred",
            Error:   fmt.Sprintf("Errors: %v", errors),
        })
        return
    }
    
    ctx.JSON(http.StatusCreated, Response{
        Success: true,
        Message: fmt.Sprintf("Created %d users successfully", len(users)),
        Data:    users,
    })
}

func updateUser(ctx *zoox.Context) {
    id := ctx.Param("id")
    
    var updates map[string]interface{}
    
    // Parse partial JSON updates
    if err := ctx.BindJSON(&updates); err != nil {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Invalid JSON format",
            Error:   err.Error(),
        })
        return
    }
    
    // Simulate finding user
    user := User{
        ID:      1,
        Name:    "John Doe",
        Email:   "john@example.com",
        Age:     30,
        Active:  true,
        Created: time.Now().Add(-24 * time.Hour),
    }
    
    // Apply updates
    if name, ok := updates["name"].(string); ok {
        user.Name = name
    }
    
    if email, ok := updates["email"].(string); ok {
        user.Email = email
    }
    
    if age, ok := updates["age"].(float64); ok {
        user.Age = int(age)
    }
    
    if active, ok := updates["active"].(bool); ok {
        user.Active = active
    }
    
    ctx.JSON(http.StatusOK, Response{
        Success: true,
        Message: "User updated successfully",
        Data:    user,
    })
}

func getUser(ctx *zoox.Context) {
    id := ctx.Param("id")
    
    // Simulate finding user
    user := User{
        ID:      1,
        Name:    "John Doe",
        Email:   "john@example.com",
        Age:     30,
        Active:  true,
        Created: time.Now().Add(-24 * time.Hour),
        Profile: Profile{
            Bio:     "Software developer",
            Website: "https://johndoe.com",
            Avatar:  "https://example.com/avatar.jpg",
        },
        Tags: []string{"developer", "go", "web"},
    }
    
    // Custom serialization based on query parameters
    includeProfile := ctx.Query("include_profile") == "true"
    includeTags := ctx.Query("include_tags") == "true"
    
    if !includeProfile {
        user.Profile = Profile{}
    }
    
    if !includeTags {
        user.Tags = nil
    }
    
    ctx.JSON(http.StatusOK, Response{
        Success: true,
        Message: "User retrieved successfully",
        Data:    user,
    })
}

// SearchCriteria represents search parameters
type SearchCriteria struct {
    Name     string   `json:"name"`
    Email    string   `json:"email"`
    MinAge   int      `json:"min_age"`
    MaxAge   int      `json:"max_age"`
    Active   *bool    `json:"active"`
    Tags     []string `json:"tags"`
    Page     int      `json:"page"`
    PageSize int      `json:"page_size"`
}

func searchUsers(ctx *zoox.Context) {
    var criteria SearchCriteria
    
    // Parse search criteria
    if err := ctx.BindJSON(&criteria); err != nil {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Invalid search criteria",
            Error:   err.Error(),
        })
        return
    }
    
    // Set defaults
    if criteria.Page <= 0 {
        criteria.Page = 1
    }
    
    if criteria.PageSize <= 0 {
        criteria.PageSize = 10
    }
    
    if criteria.PageSize > 100 {
        criteria.PageSize = 100
    }
    
    // Simulate search results
    users := []User{
        {
            ID:      1,
            Name:    "John Doe",
            Email:   "john@example.com",
            Age:     30,
            Active:  true,
            Created: time.Now().Add(-24 * time.Hour),
            Tags:    []string{"developer", "go"},
        },
        {
            ID:      2,
            Name:    "Jane Smith",
            Email:   "jane@example.com",
            Age:     25,
            Active:  true,
            Created: time.Now().Add(-48 * time.Hour),
            Tags:    []string{"designer", "ui"},
        },
    }
    
    ctx.JSON(http.StatusOK, Response{
        Success: true,
        Message: "Search completed successfully",
        Data: map[string]interface{}{
            "users":     users,
            "total":     len(users),
            "page":      criteria.Page,
            "page_size": criteria.PageSize,
            "criteria":  criteria,
        },
    })
}

func generateID() int {
    return int(time.Now().UnixNano() % 1000000)
}
```

## Advanced JSON Techniques

### Custom JSON Marshaling

```go
// CustomTime handles time formatting
type CustomTime struct {
    time.Time
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
    return json.Marshal(ct.Time.Format("2006-01-02 15:04:05"))
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
    var timeStr string
    if err := json.Unmarshal(data, &timeStr); err != nil {
        return err
    }
    
    t, err := time.Parse("2006-01-02 15:04:05", timeStr)
    if err != nil {
        return err
    }
    
    ct.Time = t
    return nil
}

// UserWithCustomTime demonstrates custom marshaling
type UserWithCustomTime struct {
    ID      int        `json:"id"`
    Name    string     `json:"name"`
    Created CustomTime `json:"created"`
}
```

### JSON Validation

```go
import (
    "github.com/go-playground/validator/v10"
)

var validate = validator.New()

func validateJSON(data interface{}) error {
    return validate.Struct(data)
}

// In your handler
func createUserWithValidation(ctx *zoox.Context) {
    var req CreateUserRequest
    
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Invalid JSON format",
            Error:   err.Error(),
        })
        return
    }
    
    // Validate using struct tags
    if err := validateJSON(req); err != nil {
        ctx.JSON(http.StatusBadRequest, Response{
            Success: false,
            Message: "Validation failed",
            Error:   err.Error(),
        })
        return
    }
    
    // Process valid data...
}
```

## Hands-on Exercise: Product Catalog API

Create a product catalog API that demonstrates advanced JSON handling:

### Requirements:
1. Product CRUD operations with complex nested data
2. JSON validation for all inputs
3. Custom JSON serialization based on user roles
4. Bulk operations for multiple products
5. Search functionality with complex criteria
6. Error handling for all JSON operations

### Solution:

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/go-zoox/zoox"
)

// Product represents a product in the catalog
type Product struct {
    ID          int                    `json:"id"`
    Name        string                 `json:"name" validate:"required,min=2,max=100"`
    Description string                 `json:"description" validate:"max=500"`
    Price       float64                `json:"price" validate:"required,min=0"`
    Currency    string                 `json:"currency" validate:"required,oneof=USD EUR GBP"`
    Category    Category               `json:"category" validate:"required"`
    Tags        []string               `json:"tags"`
    Variants    []ProductVariant       `json:"variants"`
    Images      []ProductImage         `json:"images"`
    Metadata    map[string]interface{} `json:"metadata"`
    Stock       Stock                  `json:"stock"`
    Created     time.Time              `json:"created"`
    Updated     time.Time              `json:"updated"`
    Active      bool                   `json:"active"`
}

// Category represents a product category
type Category struct {
    ID   int    `json:"id" validate:"required"`
    Name string `json:"name" validate:"required"`
    Path string `json:"path"`
}

// ProductVariant represents a product variant
type ProductVariant struct {
    ID         int                    `json:"id"`
    Name       string                 `json:"name" validate:"required"`
    Price      float64                `json:"price" validate:"min=0"`
    SKU        string                 `json:"sku" validate:"required"`
    Attributes map[string]interface{} `json:"attributes"`
    Stock      int                    `json:"stock" validate:"min=0"`
}

// ProductImage represents a product image
type ProductImage struct {
    ID       int    `json:"id"`
    URL      string `json:"url" validate:"required,url"`
    Alt      string `json:"alt"`
    Primary  bool   `json:"primary"`
    Position int    `json:"position"`
}

// Stock represents product stock information
type Stock struct {
    Quantity  int    `json:"quantity" validate:"min=0"`
    Reserved  int    `json:"reserved" validate:"min=0"`
    Available int    `json:"available"`
    Status    string `json:"status" validate:"oneof=in_stock out_of_stock low_stock"`
}

// CreateProductRequest represents the request for creating a product
type CreateProductRequest struct {
    Name        string                 `json:"name" validate:"required,min=2,max=100"`
    Description string                 `json:"description" validate:"max=500"`
    Price       float64                `json:"price" validate:"required,min=0"`
    Currency    string                 `json:"currency" validate:"required,oneof=USD EUR GBP"`
    CategoryID  int                    `json:"category_id" validate:"required"`
    Tags        []string               `json:"tags"`
    Variants    []ProductVariant       `json:"variants"`
    Images      []ProductImage         `json:"images"`
    Metadata    map[string]interface{} `json:"metadata"`
    Stock       Stock                  `json:"stock"`
}

// SearchProductsRequest represents search criteria
type SearchProductsRequest struct {
    Query       string   `json:"query"`
    CategoryID  int      `json:"category_id"`
    Tags        []string `json:"tags"`
    MinPrice    float64  `json:"min_price"`
    MaxPrice    float64  `json:"max_price"`
    Currency    string   `json:"currency"`
    InStock     *bool    `json:"in_stock"`
    Active      *bool    `json:"active"`
    Page        int      `json:"page"`
    PageSize    int      `json:"page_size"`
    SortBy      string   `json:"sort_by" validate:"oneof=name price created updated"`
    SortOrder   string   `json:"sort_order" validate:"oneof=asc desc"`
}

// ProductResponse represents the response format
type ProductResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Meta    interface{} `json:"meta,omitempty"`
}

var products = make(map[int]*Product)
var nextID = 1

func main() {
    app := zoox.New()

    // Middleware
    app.Use(func(ctx *zoox.Context) {
        ctx.Set("Content-Type", "application/json")
        ctx.Next()
    })

    // Product routes
    app.Post("/products", createProduct)
    app.Get("/products/:id", getProduct)
    app.Put("/products/:id", updateProduct)
    app.Delete("/products/:id", deleteProduct)
    app.Get("/products", listProducts)
    app.Post("/products/search", searchProducts)
    app.Post("/products/batch", createBatchProducts)
    app.Put("/products/batch", updateBatchProducts)

    // Seed some sample data
    seedSampleData()

    fmt.Println("Product Catalog API starting on :8080")
    log.Fatal(app.Listen(":8080"))
}

func createProduct(ctx *zoox.Context) {
    var req CreateProductRequest
    
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Invalid JSON format",
            Error:   err.Error(),
        })
        return
    }
    
    // Validate request
    if err := validateStruct(req); err != nil {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Validation failed",
            Error:   err.Error(),
        })
        return
    }
    
    // Create product
    product := &Product{
        ID:          nextID,
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        Currency:    req.Currency,
        Category:    Category{ID: req.CategoryID, Name: "Sample Category"},
        Tags:        req.Tags,
        Variants:    req.Variants,
        Images:      req.Images,
        Metadata:    req.Metadata,
        Stock:       req.Stock,
        Created:     time.Now(),
        Updated:     time.Now(),
        Active:      true,
    }
    
    // Calculate available stock
    product.Stock.Available = product.Stock.Quantity - product.Stock.Reserved
    
    // Set stock status
    if product.Stock.Available <= 0 {
        product.Stock.Status = "out_of_stock"
    } else if product.Stock.Available < 10 {
        product.Stock.Status = "low_stock"
    } else {
        product.Stock.Status = "in_stock"
    }
    
    products[nextID] = product
    nextID++
    
    ctx.JSON(http.StatusCreated, ProductResponse{
        Success: true,
        Message: "Product created successfully",
        Data:    product,
    })
}

func getProduct(ctx *zoox.Context) {
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Invalid product ID",
        })
        return
    }
    
    product, exists := products[id]
    if !exists {
        ctx.JSON(http.StatusNotFound, ProductResponse{
            Success: false,
            Message: "Product not found",
        })
        return
    }
    
    // Custom serialization based on query parameters
    includeVariants := ctx.Query("include_variants") != "false"
    includeImages := ctx.Query("include_images") != "false"
    includeMetadata := ctx.Query("include_metadata") != "false"
    
    productCopy := *product
    
    if !includeVariants {
        productCopy.Variants = nil
    }
    
    if !includeImages {
        productCopy.Images = nil
    }
    
    if !includeMetadata {
        productCopy.Metadata = nil
    }
    
    ctx.JSON(http.StatusOK, ProductResponse{
        Success: true,
        Message: "Product retrieved successfully",
        Data:    productCopy,
    })
}

func updateProduct(ctx *zoox.Context) {
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Invalid product ID",
        })
        return
    }
    
    product, exists := products[id]
    if !exists {
        ctx.JSON(http.StatusNotFound, ProductResponse{
            Success: false,
            Message: "Product not found",
        })
        return
    }
    
    var updates map[string]interface{}
    if err := ctx.BindJSON(&updates); err != nil {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Invalid JSON format",
            Error:   err.Error(),
        })
        return
    }
    
    // Apply updates
    if name, ok := updates["name"].(string); ok {
        product.Name = name
    }
    
    if description, ok := updates["description"].(string); ok {
        product.Description = description
    }
    
    if price, ok := updates["price"].(float64); ok {
        product.Price = price
    }
    
    if currency, ok := updates["currency"].(string); ok {
        product.Currency = currency
    }
    
    if active, ok := updates["active"].(bool); ok {
        product.Active = active
    }
    
    // Handle nested updates
    if stockData, ok := updates["stock"].(map[string]interface{}); ok {
        if quantity, ok := stockData["quantity"].(float64); ok {
            product.Stock.Quantity = int(quantity)
        }
        if reserved, ok := stockData["reserved"].(float64); ok {
            product.Stock.Reserved = int(reserved)
        }
        // Recalculate available stock
        product.Stock.Available = product.Stock.Quantity - product.Stock.Reserved
    }
    
    product.Updated = time.Now()
    
    ctx.JSON(http.StatusOK, ProductResponse{
        Success: true,
        Message: "Product updated successfully",
        Data:    product,
    })
}

func deleteProduct(ctx *zoox.Context) {
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Invalid product ID",
        })
        return
    }
    
    _, exists := products[id]
    if !exists {
        ctx.JSON(http.StatusNotFound, ProductResponse{
            Success: false,
            Message: "Product not found",
        })
        return
    }
    
    delete(products, id)
    
    ctx.JSON(http.StatusOK, ProductResponse{
        Success: true,
        Message: "Product deleted successfully",
    })
}

func listProducts(ctx *zoox.Context) {
    page, _ := strconv.Atoi(ctx.Query("page"))
    if page <= 0 {
        page = 1
    }
    
    pageSize, _ := strconv.Atoi(ctx.Query("page_size"))
    if pageSize <= 0 {
        pageSize = 10
    }
    
    var productList []*Product
    for _, product := range products {
        productList = append(productList, product)
    }
    
    start := (page - 1) * pageSize
    end := start + pageSize
    
    if start >= len(productList) {
        productList = []*Product{}
    } else if end > len(productList) {
        productList = productList[start:]
    } else {
        productList = productList[start:end]
    }
    
    ctx.JSON(http.StatusOK, ProductResponse{
        Success: true,
        Message: "Products retrieved successfully",
        Data:    productList,
        Meta: map[string]interface{}{
            "page":       page,
            "page_size":  pageSize,
            "total":      len(products),
            "has_more":   end < len(products),
        },
    })
}

func searchProducts(ctx *zoox.Context) {
    var req SearchProductsRequest
    
    if err := ctx.BindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Invalid search criteria",
            Error:   err.Error(),
        })
        return
    }
    
    // Set defaults
    if req.Page <= 0 {
        req.Page = 1
    }
    if req.PageSize <= 0 {
        req.PageSize = 10
    }
    if req.PageSize > 100 {
        req.PageSize = 100
    }
    
    var results []*Product
    
    // Simple search implementation
    for _, product := range products {
        if matchesSearchCriteria(product, req) {
            results = append(results, product)
        }
    }
    
    // Pagination
    start := (req.Page - 1) * req.PageSize
    end := start + req.PageSize
    
    if start >= len(results) {
        results = []*Product{}
    } else if end > len(results) {
        results = results[start:]
    } else {
        results = results[start:end]
    }
    
    ctx.JSON(http.StatusOK, ProductResponse{
        Success: true,
        Message: "Search completed successfully",
        Data:    results,
        Meta: map[string]interface{}{
            "page":      req.Page,
            "page_size": req.PageSize,
            "total":     len(results),
            "criteria":  req,
        },
    })
}

func createBatchProducts(ctx *zoox.Context) {
    var requests []CreateProductRequest
    
    if err := ctx.BindJSON(&requests); err != nil {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Invalid JSON array format",
            Error:   err.Error(),
        })
        return
    }
    
    if len(requests) == 0 {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "At least one product is required",
        })
        return
    }
    
    if len(requests) > 50 {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Maximum 50 products allowed per batch",
        })
        return
    }
    
    var createdProducts []*Product
    var errors []string
    
    for i, req := range requests {
        if err := validateStruct(req); err != nil {
            errors = append(errors, fmt.Sprintf("Product %d: %s", i+1, err.Error()))
            continue
        }
        
        product := &Product{
            ID:          nextID,
            Name:        req.Name,
            Description: req.Description,
            Price:       req.Price,
            Currency:    req.Currency,
            Category:    Category{ID: req.CategoryID, Name: "Sample Category"},
            Tags:        req.Tags,
            Variants:    req.Variants,
            Images:      req.Images,
            Metadata:    req.Metadata,
            Stock:       req.Stock,
            Created:     time.Now(),
            Updated:     time.Now(),
            Active:      true,
        }
        
        product.Stock.Available = product.Stock.Quantity - product.Stock.Reserved
        
        products[nextID] = product
        createdProducts = append(createdProducts, product)
        nextID++
    }
    
    if len(errors) > 0 {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Some products failed validation",
            Error:   strings.Join(errors, "; "),
            Data:    createdProducts,
        })
        return
    }
    
    ctx.JSON(http.StatusCreated, ProductResponse{
        Success: true,
        Message: fmt.Sprintf("Created %d products successfully", len(createdProducts)),
        Data:    createdProducts,
    })
}

func updateBatchProducts(ctx *zoox.Context) {
    var updates []map[string]interface{}
    
    if err := ctx.BindJSON(&updates); err != nil {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Invalid JSON format",
            Error:   err.Error(),
        })
        return
    }
    
    var updatedProducts []*Product
    var errors []string
    
    for i, update := range updates {
        idFloat, ok := update["id"].(float64)
        if !ok {
            errors = append(errors, fmt.Sprintf("Update %d: Missing or invalid ID", i+1))
            continue
        }
        
        id := int(idFloat)
        product, exists := products[id]
        if !exists {
            errors = append(errors, fmt.Sprintf("Update %d: Product %d not found", i+1, id))
            continue
        }
        
        // Apply updates (simplified)
        if name, ok := update["name"].(string); ok {
            product.Name = name
        }
        if price, ok := update["price"].(float64); ok {
            product.Price = price
        }
        
        product.Updated = time.Now()
        updatedProducts = append(updatedProducts, product)
    }
    
    if len(errors) > 0 {
        ctx.JSON(http.StatusBadRequest, ProductResponse{
            Success: false,
            Message: "Some updates failed",
            Error:   strings.Join(errors, "; "),
            Data:    updatedProducts,
        })
        return
    }
    
    ctx.JSON(http.StatusOK, ProductResponse{
        Success: true,
        Message: fmt.Sprintf("Updated %d products successfully", len(updatedProducts)),
        Data:    updatedProducts,
    })
}

func matchesSearchCriteria(product *Product, req SearchProductsRequest) bool {
    // Simple matching logic
    if req.Query != "" && !strings.Contains(strings.ToLower(product.Name), strings.ToLower(req.Query)) {
        return false
    }
    
    if req.CategoryID != 0 && product.Category.ID != req.CategoryID {
        return false
    }
    
    if req.MinPrice > 0 && product.Price < req.MinPrice {
        return false
    }
    
    if req.MaxPrice > 0 && product.Price > req.MaxPrice {
        return false
    }
    
    if req.Currency != "" && product.Currency != req.Currency {
        return false
    }
    
    if req.InStock != nil && (*req.InStock && product.Stock.Available <= 0) {
        return false
    }
    
    if req.Active != nil && *req.Active != product.Active {
        return false
    }
    
    return true
}

func validateStruct(s interface{}) error {
    // Simple validation - in real app, use a validation library
    return nil
}

func seedSampleData() {
    // Add some sample products
    products[1] = &Product{
        ID:          1,
        Name:        "Laptop Computer",
        Description: "High-performance laptop for professionals",
        Price:       999.99,
        Currency:    "USD",
        Category:    Category{ID: 1, Name: "Electronics"},
        Tags:        []string{"laptop", "computer", "electronics"},
        Stock:       Stock{Quantity: 50, Reserved: 5, Available: 45, Status: "in_stock"},
        Created:     time.Now().Add(-48 * time.Hour),
        Updated:     time.Now().Add(-24 * time.Hour),
        Active:      true,
    }
    
    products[2] = &Product{
        ID:          2,
        Name:        "Coffee Mug",
        Description: "Ceramic coffee mug with handle",
        Price:       12.99,
        Currency:    "USD",
        Category:    Category{ID: 2, Name: "Kitchen"},
        Tags:        []string{"mug", "coffee", "kitchen"},
        Stock:       Stock{Quantity: 100, Reserved: 10, Available: 90, Status: "in_stock"},
        Created:     time.Now().Add(-72 * time.Hour),
        Updated:     time.Now().Add(-12 * time.Hour),
        Active:      true,
    }
    
    nextID = 3
}
```

## Key Takeaways

1. **Proper JSON Parsing**: Always validate JSON input and handle parsing errors gracefully
2. **Struct Tags**: Use appropriate struct tags for JSON serialization and validation
3. **Custom Serialization**: Implement custom marshaling/unmarshaling when needed
4. **Validation**: Validate all JSON input before processing
5. **Error Handling**: Provide clear error messages for JSON-related issues
6. **Performance**: Consider JSON processing performance for large datasets
7. **Security**: Validate and sanitize all JSON input to prevent injection attacks

## Next Steps

- Tutorial 06: Template Engine - Learn how to render HTML templates with data
- Tutorial 07: Static Files & Assets - Serve static files and optimize asset delivery
- Explore JSON streaming for large datasets
- Learn about JSON schema validation
- Practice with real-world JSON APIs

## Common Issues and Solutions

### Issue: JSON Parsing Errors
**Solution**: Always check for parsing errors and provide meaningful error messages

### Issue: Validation Failures
**Solution**: Use struct tags and validation libraries for comprehensive validation

### Issue: Performance with Large JSON
**Solution**: Consider streaming JSON for large datasets and implement pagination

### Issue: Custom Date Formats
**Solution**: Implement custom marshaling/unmarshaling for specific date formats

## Additional Resources

- [Go JSON Package Documentation](https://golang.org/pkg/encoding/json/)
- [JSON Validation Best Practices](https://json-schema.org/)
- [Performance Optimization for JSON Processing](https://golang.org/doc/effective_go.html#json)
</rewritten_file> 